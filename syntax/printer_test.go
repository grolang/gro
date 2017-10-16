// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"os"
	"testing"
)

func TestPrint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	asts, err := ParseFile(*src_, nil, nil, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(asts) != 1 {
		t.Error(fmt.Sprintf("More than one file returned from parse of %s.", *src_))
	}
	for _, ast:= range asts {
		Fprint(os.Stdout, ast, true)
		fmt.Println()
	}
}

func TestPrintString(t *testing.T) {
	for _, want := range []string{
		"package p",
		"package p; type _ = int; type T1 = struct{}; type ( _ = *struct{}; T2 = float32 )",
		// TODO(gri) expand
	} {
		asts, err:= ParseBytes("dud", nil, []byte(want), nil, nil, 0, nil)
		if err != nil {
			t.Error(err)
			continue
		}
		if len(asts) != 1 {
			t.Error(fmt.Sprintf("More than one file returned from parse of supplied string:\n%s", want))
		}
		for _, ast:= range asts {
			if got := String(ast); got != want {
				t.Errorf("%q: got %q", want, got)
			}
		}
	}
}

