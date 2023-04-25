package gordian

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"sync/atomic"
)

// circular buffer underlying PipeReader and PipeWriter
type pipe struct {
	w   atomic.Int32
	buf []byte

	mx     sync.Mutex
	r      int
	closed bool
}

// Pipe creates a synchronous in-memory pipe.
// It can be used to connect multiple passes in the Gordian program.
//
// It relies internally on a circular buffer to pass data between the Gordian passes without copy.
//
// It is safe to call Read and Write in parallel with each other or with Close.
// Parallel calls to Read and parallel calls to Write are not safe.
func Pipe() (*Reader, *Writer) {
	p := &pipe{buf: make([]byte, 80_000)}
	wc := &sync.Cond{L: &p.mx}
	wc.L.Lock()
	return &Reader{p: p, wc: wc, MaxSize: 40_000}, &Writer{p: p, wc: wc}
}

type Reader struct {
	MaxSize int

	p   *pipe
	i   int
	err error
	wc  *sync.Cond
}

func ReadLine(r *Reader) bool {
	r.Release(r.i)

	i := bytes.IndexByte(r.Window(), '\n')
	for i == -1 {
		if !r.Next(1024) {
			return false
		}
		i = bytes.IndexByte(r.Window(), '\n')
	}
	r.i = i
	return true
}

var ErrShortRead = errors.New("read too short")

// Window returns the current window viewed by the reader
// Bytes returned by Window are only valid between a call to Next, and a call to Release
func (r *Reader) Window() []byte { return r.p.buf[r.p.r : r.p.r+r.i] }

// Release frees up n bytes in the underlying pipe.
// Calling Release invalidates bytes returned by Window.
func (r *Reader) Release(n int) {
	r.p.r += n
	r.i -= n
	r.wc.Signal()
}

// Next will try to read up to n bytes from the corresponding writer.
// It might read less bytes if [MaxSize] is reached.
// It return false if the corresponding writer is closed, and all bytes have been read.
func (r *Reader) Next(n int) bool {
	// concurrency note: up can only increase after the load, since we hold the lock
	// so, even if racy, the comparison is still valid
	if up := int(r.p.w.Load()); up-(r.p.r+r.i) >= n {
		r.i += n
		return true
	}

	for !r.p.closed {
		r.wc.Wait()

		up := int(r.p.w.Load())
		switch {
		case up-r.p.r > r.i+n:
			r.i += n
			return true
		case up-r.p.r > r.MaxSize:
			r.i = r.MaxSize
			return true
		}
	}
	r.i = int(r.p.w.Load()) - r.p.r
	return r.i > 0
}

func (r *Reader) Read(dt []byte) (int, error) {
	if len(r.Window()) < len(dt) {
		more := r.Next(len(dt) - len(r.Window()))
		if !more {
			return 0, io.EOF
		}
	}

	s := copy(dt, r.Window())
	r.Release(s)
	return s, nil
}

type Writer struct {
	p  *pipe
	wc *sync.Cond
}

// BUG(rdo) there is a deadlock condition here: if the Write needs to go to the slow path, and wait for the Read to happen,
// and that there are not enough Read to release space, we can deadlock.
func (w *Writer) Write(dt []byte) (int, error) {
	up := int(w.p.w.Load())
	if len(dt) <= len(w.p.buf)-up {
		copy(w.p.buf[up:], dt)
		w.p.w.Swap(int32(up + len(dt)))
		w.wc.Signal() // no-op if no one is listening
		return len(dt), nil
	}

	w.wc.L.Lock()
	for {
		copy(w.p.buf, w.p.buf[w.p.r:])
		up -= w.p.r
		w.p.r = 0
		w.p.w.Swap(int32(up))
		if len(dt) <= len(w.p.buf)-up {
			break
		} else {
			w.wc.Wait()
		}
	}

	copy(w.p.buf[up:], dt)
	w.p.w.Swap(int32(up + len(dt)))
	w.wc.Signal()
	w.wc.L.Unlock()
	return len(dt), nil
}

func (w *Writer) Close() error {
	w.wc.L.Lock()
	w.p.closed = true
	w.wc.Signal()
	w.wc.L.Unlock()

	return nil
}
