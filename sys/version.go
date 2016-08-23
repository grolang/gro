// Copyright 2011 The Go and Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"fmt"
	"os"
	"runtime"
)

var cmdVersion = &Command{
	Run:       runVersion,
	UsageLine: "version",
	Short:     "print Gro version",
	Long:      `
Version prints the Gro version.

	`,
}

func runVersion(cmd *Command, cmds []*Command, args []string) {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "usage: gro version\n\nToo many arguments given.\n")
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "Gro version %s, running on Go version %s %s/%s\n", versionNo, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

