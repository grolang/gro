// Copyright 2016 The Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"github.com/grolang/gro/tutor"
	"github.com/grolang/gro/tutor/basic"
)

//================================================================================
type LearnState struct {
	currTut    int
	currPage   int
}

var learnState LearnState

//================================================================================
func (s *LearnState) setTut (scr string) bool {
	for n, tut := range tutorials {
		if tut.Name == scr {
			s.currTut = n
			return true
		}
	}
	return false
}

//================================================================================
func (s *LearnState) printScreen (offset int) (finished bool) {
	s.currPage += offset
	if s.currPage >= len(tutorials[s.currTut].Pages) {
		fmt.Fprintf(os.Stderr, "\nNo more tutorial pages left.\nContinue practising in the Repl!\n")
		return true
	} else if s.currPage < 0 {
		s.currPage = 0
		fmt.Fprintf(os.Stderr, tutorials[s.currTut].Name, "page", 1)
		fmt.Fprintf(os.Stderr, tutorials[s.currTut].Pages[0].EnglishPage())
		return false
	} else {
		fmt.Fprintf(os.Stderr, tutorials[s.currTut].Name, "page", s.currPage+1)
		fmt.Fprintf(os.Stderr, tutorials[s.currTut].Pages[s.currPage].EnglishPage())
		return false
	}
}

//================================================================================
var cmdLearn = &Command{
	Run:       runLearn,
	UsageLine: "learn [tutorial-name]",
	Short:     "give an interactive tutorial on the Gro language and tool",
	Long:      `
Learn gives an interactive tutorial on the Gro language and tool.

From within the learn mode, type a valid statement, or:

	后      advance a page
	前      go back a page
	出      exit learn mode

The tutorials available are:
{{range .}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}

`,
}

//================================================================================
func runLearn(cmd *Command, cmds []*Command, args []string) {
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: gro learn tutorial-name\n\nWrong number of arguments given.\n")
		os.Exit(2)
	}
	procRepl(cmd, cmds, args)
}

// ============================================================================
func AddTutorial (t *tutor.Tutorial) {
	tutorials = append(tutorials, t)
}

// ============================================================================
// Tutorials lists the available tutorials.
// The order here is the order in which they are printed by 'gro learn'.
var tutorials = []*tutor.Tutorial{
	basic.TutIntro,
	basic.TutStmts,
	basic.TutTodo,
}

func Tutorials() []*tutor.Tutorial {
	return tutorials
}

// ============================================================================
func printTutorials(w io.Writer) {
	bw := bufio.NewWriter(w)
	MakeTemplate(bw, cmdLearn.Long, tutorials)
	bw.Flush()
}

//================================================================================

