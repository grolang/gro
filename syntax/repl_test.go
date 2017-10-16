// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"
	"testing"
)

func TestRepl(t *testing.T){
	for _, tst:= range replTests {
		got, err:= ParseReplCmd(tst.pin, tst.sin)
		if err != nil {
			t.Error(fmt.Sprintf("Test %d: Error received: %s", tst.num, err))
			continue
		}
		if got != tst.out {
			t.Errorf("Test %d: expected and received output not the same.\n\n#### Expected:\n%s\n\n#### Received:\n%s\n\n",
				tst.num, tst.out, got)
		}
	}
}

var replTests = []struct{
	num   int
	pin   string
	sin   []string
	out   string
}{

//--------------------------------------------------------------------------------
	{
		num:  10,
		pin:  `var(
	a = 1
)
const(
)
`,
		sin:  []string{`do a:= 7`},
// - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
		out:  `import(
	"io/ioutil"
	"os"
)
_= ioutil.WriteFile("temp/repl-var.gro", []byte("var(\na=" + fmt.Sprintf("%#v", a) + "\nb=" + fmt.Sprintf("%#v", b) + "\n)\nconst(\na=" + fmt.Sprintf("%#v", a) + "\nb=" + fmt.Sprintf("%#v", b) + "\n)\n"), os.ModeExclusive)`,
	},

//--------------------------------------------------------------------------------
}

