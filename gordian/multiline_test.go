package main

import (
	"strings"
	"testing"
)

func TestMultiLine(t *testing.T) {
	cases := []struct{ in, out string }{
		{`****** host/examplenet, 2020:01:11 11:21:00
 application message!!
 string1 : rec_local_disk_usage_control.handle_correct_ioc_status`, `****** host/examplenet, 2020:01:11 11:21:00 application message!! string1 : rec_local_disk_usage_control.handle_correct_ioc_status`},
		{`***** another one bytes the dust *****
-----> block info  <------
-----> class : LOCAL
-----> host  : examplehost
--------------------------`, `***** another one bytes the dust ***** -----> block info  <------ -----> class : LOCAL -----> host  : examplehost --------------------------`},
	}

	for _, c := range cases {
		var m GroupML
		if err := m.Init(`\*\*\*\*\*`); err != nil {
			t.Fatal(err)
		}

		var out strings.Builder
		if err := m.Pipe(strings.NewReader(c.in), &out); err != nil {
			t.Fatal(err)
		}

		if out.String() != c.out {
			t.Errorf("\nwant: %s\n got: %s", c.out, out.String())
		}
	}
}
