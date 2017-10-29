// +build ignore

package main

import (
	fmt "fmt"
	sys "github.com/grolang/gro/sys"
)

func init() {
	fmt.Println("'Hello, world!' from executeInside.gro")
}

func init() {
	sys.Execute("testdata/sayhi.gro")
}

func main() {}