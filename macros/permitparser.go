// Copyright 2017 The Gro Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package macros

//--------------------------------------------------------------------------------
func GroSys(p PermitParser, lits []string) {
}

//--------------------------------------------------------------------------------
type PermitParser interface {
	SetPermit(string)
	UnsetPermit(string)
	IsPermit(string) bool
}

func InitBlacklist(p PermitParser, lits []string) {
	for _, s:= range lits {
		switch s {
		default:
			p.UnsetPermit(s)
		case "package":
			p.UnsetPermit("package")
			p.UnsetPermit("internal")
		case "section":
			p.UnsetPermit("section")
			p.UnsetPermit("main")
			p.UnsetPermit("testcode")
		case "if":
			p.UnsetPermit("if")
			p.UnsetPermit("else")
		case "switch":
			p.UnsetPermit("switch")
			p.UnsetPermit("fallthrough")
			if ! p.IsPermit("select") {
				p.UnsetPermit("case")
				p.UnsetPermit("default")
			}
			if ! p.IsPermit("for") && ! p.IsPermit("select") {
				p.UnsetPermit("break")
			}
		case "select":
			p.UnsetPermit("select")
			if ! p.IsPermit("switch") {
				p.UnsetPermit("case")
				p.UnsetPermit("default")
			}
			if ! p.IsPermit("for") && ! p.IsPermit("switch") {
				p.UnsetPermit("break")
			}
		case "for":
			p.UnsetPermit("for")
			p.UnsetPermit("range")
			p.UnsetPermit("continue")
			if ! p.IsPermit("switch") && ! p.IsPermit("select") {
				p.UnsetPermit("break")
			}
		}
	}
}

//--------------------------------------------------------------------------------
