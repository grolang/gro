// Copyright 2016-17 The Go and Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"text/template"
	"unicode"
	"unicode/utf8"
	"github.com/grolang/gro/sys"
)

const (
	versionNo = "0.7.2"
	progDesc = "managing Gro scripts"
)

//================================================================================
func main() {
	log.SetFlags(0)
	flag.Parse()
	args:= flag.Args()
	Main(args)
	os.Exit(sys.ExitStatus)
}

var exitMu sync.Mutex

func setExitStatus(n int) {
	exitMu.Lock()
	if sys.ExitStatus < n {
		sys.ExitStatus = n
	}
	exitMu.Unlock()
}

//--------------------------------------------------------------------------------
var (
	cpuprofile string
)

func addBuildFlags(cmd *Command) {
	cmd.Flag = *flag.NewFlagSet("", flag.ContinueOnError) // needs to be reset each time for test suite
	cmd.Flag.BoolVar(&sys.WantMsgs, "v", false, "print steps as they are executed")
	cmd.Flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to this file")
}

//================================================================================
func Main(args []string) {
	sys.ExitStatus = 0 // needs to be reset each time for test suite
	if len(args) == 0 {
		printUsage()
		return
	} else if args[0] == "help" {
		help(args[1:])
		return
	}

	// Diagnose common mistake: GOPATH==GOROOT.
	// This setting is equivalent to not setting GOPATH at all,
	// which is not what most people want when they do it.
	goroot:= runtime.GOROOT()
	if gopath := os.Getenv("GOPATH"); gopath == goroot {
		fmt.Fprintf(sys.Stderr, "%s: warning: GOPATH set to GOROOT (%s) has no effect\n", sys.ProgName, gopath)
	} else {
		for _, p := range filepath.SplitList(gopath) {
			// Note: using HasPrefix instead of Contains because a ~ can appear
			// in the middle of directory elements, such as /tmp/git-1.8.2~rc3
			// or C:\PROGRA~1. Only ~ as a path prefix has meaning to the shell.
			if strings.HasPrefix(p, "~") {
				fmt.Fprintf(sys.Stderr, "%s: GOPATH entry cannot start with shell metacharacter '~': %q\n", sys.ProgName, p)
				setExitStatus(2)
				return
			}
			if !filepath.IsAbs(p) {
				fmt.Fprintf(sys.Stderr, "%s: GOPATH entry is relative; must be absolute path: %q.\nRun 'go help gopath' for usage.\n", sys.ProgName, p)
				setExitStatus(2)
				return
			}
		}
	}
	if fi, err := os.Stat(goroot); err != nil || !fi.IsDir() {
		fmt.Fprintf(sys.Stderr, "%s: cannot find GOROOT directory: %v\n", sys.ProgName, goroot)
		setExitStatus(2)
		return
	}

	for _, cmd := range Commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			if cpuprofile != "" {
				f, err := os.Create(cpuprofile)
				if err != nil {
					fmt.Fprintf(sys.Stderr, "creating cpu profile: %s\n", err)
					setExitStatus(2)
					return
				}
				defer f.Close()
				pprof.StartCPUProfile(f)
				defer pprof.StopCPUProfile()
			}
			cmd.Flag.Usage = func() {
				cmd.Usage()
				setExitStatus(2)
			}
			addBuildFlags(cmd)
			cmd.Flag.Parse(args[1:])
			args = cmd.Flag.Args()
			cmd.Run(args...)
			return
		}
	}

	fmt.Fprintf(sys.Stderr, "%s: unknown subcommand %q\nRun 'gro help' for usage.\n", sys.ProgName, args[0])
	setExitStatus(2)
}

//================================================================================
func version(args ...string) {
	if len(args) != 0 {
		fmt.Fprintf(sys.Stderr, "usage: gro version\n\nToo many arguments given.\n")
		setExitStatus(2)
		return
	}
	fmt.Fprintf(sys.Stderr, "Gro version %s, running on Go version %s for OS:%s and arch:%s\n",
		versionNo, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

//================================================================================

// A Command is an implementation of a gro command
// like gro prepare or gro execute.
type Command struct {
	// Run runs the command.
	// The args are the arguments after the command name.
	Run func(args ...string)

	// UsageLine is the one-line usage message.
	// The first word in the line is taken to be the command name.
	UsageLine string

	// Short is the short description shown in the 'gro help' output.
	Short string

	// Long is the long message shown in the 'gro help <this-command>' output.
	Long string

	// Flag is a set of flags specific to this command.
	Flag flag.FlagSet
}

//--------------------------------------------------------------------------------

// Name returns the command's name: the first word in the usage line.
func (c *Command) Name() string {
	name := c.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

//--------------------------------------------------------------------------------

func (c *Command) Usage() {
	fmt.Fprintf(sys.Stderr, "usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(sys.Stderr, "%s\n", strings.TrimSpace(c.Long))
}

//--------------------------------------------------------------------------------

// Runnable reports whether the command can be run; otherwise
// it is a documentation pseudo-command such as importpath.
func (c *Command) Runnable() bool {
	return c.Run != nil
}

// ============================================================================
// Commands lists the available commands and help topics.
// The order here is the order in which they are printed by 'gro help'.
var Commands = []*Command {
	cmdPrepare,
	cmdExecute,
	cmdVersion,

	helpFlags,
}

//--------------------------------------------------------------------------------
var cmdPrepare = &Command{
	Run:       sys.Prepare,
	UsageLine: "prepare [flags] [path ...]",
	Short:     "generate the go files",
	Long: `
Prepare generates formatted Go programs from Gro scripts.
It uses the same whitespace as gofmt.

Without an explicit path, it processes the standard input and prints the reformatted source to the standard output.
Given a file, it operates on that file; given a directory, it operates on all .gro files in that directory, recursively.
(Files starting with a period are ignored.)
It then prints the generated Go source to the output files as determined by the Gro source.

`,
}

//--------------------------------------------------------------------------------
var cmdExecute = &Command{
	Run:       sys.Execute,
	UsageLine: "execute [flags] [path ...]",
	Short:     "generate the go files then run the main func",
	Long: `
Execute first prepares the Gro scripts, then runs package main function main()
on the go file with the same name as the first gro file.
See gro help prepare.

`,
}

//--------------------------------------------------------------------------------
var cmdVersion = &Command{
	Run:       version,
	UsageLine: "version",
	Short:     "print Gro version",
	Long:      `
Version prints the Gro version.

`,
}

//--------------------------------------------------------------------------------
var helpFlags = &Command{
	UsageLine: "flags",
	Short:     "flags used in Gro",
	Long: `
The flags common to both prepare and execute are:

	-cpuprofile filename
		Write cpu profile to the specified file.
	-v
		Print steps as they are executed.

`,
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

// MakeTemplate executes the given template text on data, writing the result to w.
func MakeTemplate(w io.Writer, text string, data interface{}) {
	t := template.New("top")
	t.Funcs(template.FuncMap{
		"trim": strings.TrimSpace,
		"capitalize": capitalize,
		"progname": ProgName,
		"progdesc": ProgDesc,
		"progvers": ProgVers,
		"suffix":   Suffix,
	})
	template.Must(t.Parse(text))
	ew := &errWriter{w: w}
	err := t.Execute(ew, data)
	if ew.err != nil {
		// I/O error writing. Ignore write on closed pipe.
		if strings.Contains(ew.err.Error(), "pipe") {
			setExitStatus(1)
			return
		}
		log.Printf("writing output: %v", ew.err)
		setExitStatus(1)
		return
	}
	if err != nil {
		panic(err)
	}
}

//--------------------------------------------------------------------------------
func ProgName() string { return sys.ProgName }
func ProgDesc() string { return progDesc }
func ProgVers() string { return versionNo }
func Suffix() string { return sys.Suffix }

//--------------------------------------------------------------------------------
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToTitle(r)) + s[n:]
}

// ============================================================================
func printUsage() {
	bw := bufio.NewWriter(sys.Stderr)
	MakeTemplate(bw, usageTemplate, Commands)
	bw.Flush()
}

//--------------------------------------------------------------------------------
const usageTemplate = `
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
// help implements the 'help' command.
func help(args []string) {
	if len(args) == 0 {
		printUsage()
		return
	}
	if len(args) > 2 {
		fmt.Fprintf(sys.Stderr, "usage: %s help command\n\nToo many arguments given.\n", sys.ProgName)
		setExitStatus(2)
		return
	}
	arg := args[0]
	for _, cmd := range Commands {
		if cmd.Name() == arg {
			MakeTemplate(sys.Stderr, helpTemplate, cmd)
			return
		}
	}
	fmt.Fprintf(sys.Stderr, "Unknown help topic %#q.  Run '%s help'.\n", arg, sys.ProgName)
	setExitStatus(2)
}

//--------------------------------------------------------------------------------
const helpTemplate = `
{{if .Runnable}}usage: {{progname}} {{.UsageLine}}

{{end}}{{.Long | trim}}

`

//================================================================================

