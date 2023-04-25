package gordian

import (
	"errors"
	"sync"
)

// circular buffer underlying PipeReader and PipeWriter
type pipe struct {
	mx     sync.Mutex
	buf    []byte
	r, w   int
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
	err error
	wc  *sync.Cond
}

var ErrShortRead = errors.New("read too short")

// Window returns the current window viewed by the reader
// Bytes returned by Window are only valid between a call to Next, and a call to Release
func (r *Reader) Window() []byte { return r.p.buf[r.p.r:r.p.w] }

// Release frees up n bytes in the underlying pipe.
// Calling Release invalidates bytes returned by Window.
func (r *Reader) Release(n int) {
	r.p.r += n
	r.wc.Signal()
}

// Next will try to read up to n bytes from the corresponding writer.
// It might read less bytes if [MaxSize] is reached.
// It return false if the corresponding writer is closed, and all bytes have been read.
func (r *Reader) Next(n int) bool {
	start := r.p.w - r.p.r
	for !r.p.closed {
		avail := r.p.w - r.p.r
		r.wc.Wait()

		switch {
		case r.p.w-r.p.r-avail == 0:
			r.err = ErrShortRead
			return false
		case r.p.w-r.p.r > r.MaxSize:
			return true
		case r.p.w-r.p.r > start+n:
			return true
		}
	}

	return r.p.w > r.p.r
}

type Writer struct {
	p  *pipe
	wc *sync.Cond
}

func (w *Writer) Write(dt []byte) {
	w.wc.L.Lock()
	for len(w.p.buf)-w.p.w < len(dt) {
		copy(w.p.buf, w.p.buf[w.p.r:])
		w.p.w -= w.p.r
		w.p.r = 0

		if len(w.p.buf)-w.p.w < len(dt) {
			w.wc.Wait()
		}
	}

	copy(w.p.buf[w.p.w:], dt)
	w.p.w += len(dt)
	w.wc.Signal()
	w.wc.L.Unlock()
}

func (w *Writer) Close() error {
	w.wc.L.Lock()
	w.p.closed = true
	w.wc.Signal()
	w.wc.L.Unlock()

	return nil
}
