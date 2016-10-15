// Copyright 2009-16 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ============================================================================
var cmdExecute = &Command{
	Run:       groExecute,
	UsageLine: "execute [flags] [path ...]",
	Short:     "generate the go files then run the main func",
	Long: `
Execute first prepares the Gro scripts, then runs package main function main().
See gro help prepare.

	`,
}

// ============================================================================
func groExecute(cmd *Command, cmds []*Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: gro execute path\n\nToo many arguments given.\n")
		os.Exit(2)
	}

	groPrepare(cmd, cmds, args)

	extLen:= len(filepath.Ext(args[0]))
	outfile:= args[0][:len(args[0])-extLen] + ".go"

	if wantMsgs {
		fmt.Fprintf(os.Stderr, ">>> running %s\n", outfile)
	}
	out, err:= exec.Command("go", "run", outfile).CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s executing %s\n%s", err, outfile, out)
		return
	}
	fmt.Fprintf(os.Stdout, "%s", out)
}

// ============================================================================

