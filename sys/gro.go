// Copyright 2009-16 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package sys provides services for the Gro command.
//
package sys

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"text/template"
	"unicode"
	"unicode/utf8"
	"github.com/grolang/gro/macro"
	"github.com/grolang/gro/scanner"
)

// ============================================================================
const (
	versionNo= "0.4.0"
)

var (
	progName = "gro"
	progDesc = "managing Gro scripts"
	progVers = "0.4.0"
	suffix = "gro"
)

var (
	wantMsgs     bool
	allErrors    bool
	cpuprofile   string
	OutFile      string
)

var predefMacros = map[rune]macro.StmtMacro{}

// ============================================================================
func AddStmtMacro(name rune, def macro.StmtMacro) {
	predefMacros[name] = def
}

// ============================================================================

func SetProgData(n, d, v string) {
	progName = n
	progDesc = d
	progVers = v
}

func ProgName() string {
	return progName
}

func ProgDesc() string {
	return progDesc
}

func ProgVers() string {
	return progVers
}

// ============================================================================

func SetSuffix(sfx string) {
	suffix = sfx
}

func Suffix() string {
	return suffix
}

// ============================================================================
func addBuildFlags(cmd *Command) {
	cmd.Flag.BoolVar(&wantMsgs, "x", false, "print steps as they are executed")
	cmd.Flag.BoolVar(&allErrors, "e", false, "report all errors (not just the first 10 on different lines)")
	cmd.Flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to this file")
	cmd.Flag.StringVar(&OutFile, "o", "", "write output to this file")
}

// ============================================================================
// An errWriter wraps a writer, recording whether a write error occurred.
type errWriter struct {
	w   io.Writer
	err error
}

func (w *errWriter) Write(b []byte) (int, error) {
	n, err := w.w.Write(b)
	if err != nil {
		w.err = err
	}
	return n, err
}

// ============================================================================
var exitStatus = 0
var exitMu sync.Mutex

func SetExitStatus(n int) {
	exitMu.Lock()
	if exitStatus < n {
		exitStatus = n
	}
	exitMu.Unlock()
}

// ============================================================================

// A Command is an implementation of a gro command
// like gro prepare or gro execute.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(cmd *Command, cmds []*Command, args []string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'gro help' output.
	Short string

	// Long is the long message shown in the 'gro help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet

	// CustomFlags indicates that the command will do its own
	// flag parsing.
	CustomFlags bool
}

// ============================================================================

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// ============================================================================

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

// ============================================================================

// Runnable reports whether the command can be run; otherwise
// it is a documentation pseudo-command such as importpath.
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// ============================================================================
var cmdGro = &Command{
	Run:       func(cmd *Command, cmds []*Command, args []string){return},
	UsageLine: "[action] [flags] [path ...]",
	Short:     "runs various utilities to complement those in the Go command",
	Long: `
The actions are:

	prepare
		generate the .go file/s
	execute
		prepare a single .gro file, then run the main() function
	repl
		start read-eval-print-loop processing
	learn
		start repl in learn mode
	version
		report which version of Gro and Go

`,
}

// ============================================================================
// Commands lists the available commands and help topics.
// The order here is the order in which they are printed by 'gro help'.
var commands = []*Command {
	cmdPrepare,
	cmdExecute,
	cmdRepl,
	cmdLearn,
	cmdVersion,

	helpFlags,
}

// ============================================================================
func AddCommand(c *Command) {
	commands = append(commands, c)
}

// ============================================================================
func Main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 {
		usage()
	} else if args[0] == "help" {
		help(args[1:])
		return
	}

	// Diagnose common mistake: GOPATH==GOROOT.
	// This setting is equivalent to not setting GOPATH at all,
	// which is not what most people want when they do it.
	if gopath := os.Getenv("GOPATH"); gopath == runtime.GOROOT() {
		fmt.Fprintf(os.Stderr, "warning: GOPATH set to GOROOT (%s) has no effect\n", gopath)
	} else {
		for _, p := range filepath.SplitList(gopath) {
			// Note: using HasPrefix instead of Contains because a ~ can appear
			// in the middle of directory elements, such as /tmp/git-1.8.2~rc3
			// or C:\PROGRA~1. Only ~ as a path prefix has meaning to the shell.
			if strings.HasPrefix(p, "~") {
				fmt.Fprintf(os.Stderr, "%s: GOPATH entry cannot start with shell metacharacter '~': %q\n", progName, p)
				os.Exit(2)
			}
			if !filepath.IsAbs(p) {
				fmt.Fprintf(os.Stderr, "%s: GOPATH entry is relative; must be absolute path: %q.\nRun 'go help gopath' for usage.\n", progName, p)
				os.Exit(2)
			}
		}
	}

	/*if fi, err := os.Stat(goroot); err != nil || !fi.IsDir() {
		fmt.Fprintf(os.Stderr, "go: cannot find GOROOT directory: %v\n", goroot)
		os.Exit(2)
	}*/

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Runnable() {
			cmd.Flag.Usage = func() { cmd.Usage() }
			addBuildFlags(cmd)
			if cmd.CustomFlags {
				args = args[1:]
			} else {
				cmd.Flag.Parse(args[1:])
				args = cmd.Flag.Args()
			}
			fullCmds:= append([]*Command {cmdGro}, commands...)
			cmd.Run(cmd, fullCmds, args)
			exit()
			return
		}
	}

	fmt.Fprintf(os.Stderr, "%s: unknown subcommand %q\nRun 'gro help' for usage.\n", progName, args[0])
	SetExitStatus(2)
	exit()
}

// ============================================================================
var usageTemplate = `
{{progname}} is a tool for {{progdesc}}.

Usage:

	{{progname}} command [arguments]

The commands are:
{{range .}}{{if .Runnable}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "{{progname}} help [command]" for more information about a command.

Additional help topics:
{{range .}}{{if not .Runnable}}
	{{.Name | printf "%-11s"}} {{.Short}}{{end}}{{end}}

Use "{{progname}} help [topic]" for more information about that topic.

`

// ============================================================================
var helpFlags = &Command{
	UsageLine: "flags",
	Short:     "flags used in Gro",
	Long: `
The flags common to both prepare and execute are:

	-cpuprofile filename
		Write cpu profile to the specified file.
	-e
		Print all (including spurious) errors.
	-x
		Print steps as they are executed.

The flag used by doc is:
	-o
		Print the command or tutorial doco to this file.

`,
}

// ============================================================================
func printUsage(w io.Writer) {
	bw := bufio.NewWriter(w)
	MakeTemplate(bw, usageTemplate, commands)
	bw.Flush()
}

// ============================================================================
func usage() {
	printUsage(os.Stderr)
	os.Exit(2)
}

// ============================================================================
var helpTemplate = `
{{if .Runnable}}usage: {{progname}} {{.UsageLine}}

{{end}}{{.Long | trim}}

`

// ============================================================================
// help implements the 'help' command.
func help(args []string) {
	if len(args) == 0 {
		printUsage(os.Stdout)
		// not exit 2: succeeded at 'gro help'.
		return
	}
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s help command\n\nToo many arguments given.\n", progName)
		os.Exit(2) // failed at 'gro help'
	}

	arg := args[0]

	for _, cmd := range commands {
		if cmd.Name() == arg {
			if arg == "learn" {
				printTutorials(os.Stdout)
			} else {
				MakeTemplate(os.Stdout, helpTemplate, cmd)
			}
			// not exit 2: succeeded at 'gro help cmd'.
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown help topic %#q.  Run '%s help'.\n", arg, progName)
	os.Exit(2) // failed at 'go help cmd'
}

// ============================================================================

// MakeTemplate executes the given template text on data, writing the result to w.
func MakeTemplate(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{
		"trim": strings.TrimSpace,
		"capitalize": capitalize,
		"progname": ProgName,
		"progdesc": ProgDesc,
		"progvers": ProgVers,
		"suffix": Suffix,
	})
	template.Must(t.Parse(text))
	ew := &errWriter{w: w}
	err := t.Execute(ew, data)
	if ew.err != nil {
		// I/O error writing. Ignore write on closed pipe.
		if strings.Contains(ew.err.Error(), "pipe") {
			os.Exit(1)
		}
		fatalf("writing output: %v", ew.err)
	}
	if err != nil {
		panic(err)
	}
}

// ============================================================================
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

// ============================================================================
var atexitFuncs []func()

func atexit(f func()) {
	atexitFuncs = append(atexitFuncs, f)
}

func exit() {
	for _, f := range atexitFuncs {
		f()
	}
	os.Exit(exitStatus)
}

func fatalf(format string, args ...interface{}) {
	errorf(format, args...)
	exit()
}

func errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
	SetExitStatus(1)
}

// ============================================================================
func Report(err error) {
	scanner.PrintError(os.Stderr, err)
	SetExitStatus(2)
}

// ============================================================================

