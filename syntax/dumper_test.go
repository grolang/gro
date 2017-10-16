// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"os"
	"testing"
)

func TestDump(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	asts, err := ParseFile(*src_, nil, nil, CheckBranches, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(asts) != 1 {
		t.Error(fmt.Sprintf("More than one file returned from parse of %s.", *src_))
	}
	for _, ast:= range asts {
		Fdump(os.Stdout, ast)
	}
}
