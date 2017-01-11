// Copyright 2011 The Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sys

import (
	"bufio"
	"fmt"
	"github.com/grolang/gro/parser"
	"github.com/grolang/gro/token"
	"io/ioutil"
	"os"
	"os/exec"
)

//================================================================================
var cmdRepl = &Command{
	Run:       runRepl,
	UsageLine: "repl",
	Short:     "run the Gro read-eval-print-loop processor",
	Long:      `
Repl runs the Gro read-eval-print-loop processor.

From within the repl, type a valid Gro statement, or:

	显      see the values of all vars
	学      enter learn mode
	出      exit the repl

	`,
}

//================================================================================
func runRepl(cmd *Command, cmds []*Command, args []string) {
	if len(args) != 0 {
		fmt.Fprintf(os.Stderr, "usage: gro repl\n\nToo many arguments given.\n")
		os.Exit(2)
	}
	procRepl(cmd, cmds, args)
}

//================================================================================
func procRepl(cmd *Command, cmds []*Command, args []string) {
	cmdName:= cmd.Name()
	holddir, _:= GetHoldDir()
	_= os.Remove(holddir + "repl-var.gro")
	scanner:= bufio.NewScanner(os.Stdin)

	if cmdName == "learn" {
		ok:= learnState.setTut(args[0])
		if !ok {
			fmt.Fprintf(os.Stderr, "No such tutorial -- repl started instead.\n")
			cmdName = "repl"
		} else {
			_ = learnState.printScreen(0) //assume all tutorials have at least one page
		}
	}

	for {
		fmt.Fprintf(os.Stderr, cmdName + "> ")
		scanner.Scan()
		inp:= scanner.Text()
		repdata, _:= ioutil.ReadFile(holddir + "repl-var.gro")

		inp = string(repdata) + "\n" + inp

		fset := token.NewFileSet()
		_, state, err := parser.ParseRepl(fset, inp, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gro repl command parse failed.\n")
			fmt.Fprintf(os.Stderr, "Errors: %v\n", err)
			os.Exit(2)
		}
		if state.CmdFound == parser.ExitCmd {
			state.CmdFound = parser.NoCmd
			break
		} else if state.CmdFound == parser.LearnCmd {
			ok:= learnState.setTut(state.Tutorial)
			if !ok {
				fmt.Fprintf(os.Stderr, "No such tutorial.\n")
			} else {
				_ = learnState.printScreen(-999)
				cmdName = "learn"
			}
			continue
		}
		if cmdName == "learn" {
			switch state.CmdFound {
			case parser.PrevCmd:
				_ = learnState.printScreen(-1)
				continue
			case parser.NextCmd:
				if learnState.printScreen(1) {
					cmdName = "repl"
				}
				continue
			}
		} else if cmdName == "repl" && state.CmdFound == parser.NextCmd {
			continue
		}

		inp = state.ImpFragment + "\n" + inp + "\n" + state.VarFragment

		fname:= "repl-pgm"
		_= ioutil.WriteFile(holddir + fname + ".gro", []byte(inp), os.ModeExclusive)
		groPrepare(cmd, cmds, []string{holddir + fname + ".gro"})

		out, _:= exec.Command("go", "run", holddir + "main.go").CombinedOutput()
		fmt.Fprintf(os.Stderr, "%s", out)
	}

}

//================================================================================

