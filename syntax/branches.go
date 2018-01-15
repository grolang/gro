// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syntax

import (
	"fmt"

	"github.com/grolang/gro/nodes"
	"github.com/grolang/gro/syntax/src"
)

// TODO(gri) consider making this part of the parser code

// checkBranches checks correct use of labels and branch
// statements (break, continue, goto) in a function body.
// It catches:
//    - misplaced breaks and continues
//    - bad labeled breaks and continues
//    - invalid, unused, duplicate, and missing labels
//    - gotos jumping over variable declarations and into blocks
func checkBranches(body *nodes.BlockStmt, errh ErrorHandler) {
	if body == nil {
		return
	}

	// scope of all labels in this body
	ls := &labelScope{errh: errh}
	fwdGotos := ls.blockBranches(nil, targets{}, nil, body.Pos(), body.List)

	// If there are any forward gotos left, no matching label was
	// found for them. Either those labels were never defined, or
	// they are inside blocks and not reachable from the gotos.
	for _, fwd := range fwdGotos {
		name := fwd.Label.Value
		if l := ls.labels[name]; l != nil {
			l.used = true // avoid "defined and not used" error
			ls.err(fwd.Label.Pos(), "goto %s jumps into block starting at %s", name, l.parent.start)
		} else {
			ls.err(fwd.Label.Pos(), "label %s not defined", name)
		}
	}

	// spec: "It is illegal to define a label that is never used."
	for _, l := range ls.labels {
		if !l.used {
			l := l.lstmt.Label
			ls.err(l.Pos(), "label %s defined and not used", l.Value)
		}
	}
}

type labelScope struct {
	errh   ErrorHandler
	labels map[string]*label // all label declarations inside the function; allocated lazily
}

type label struct {
	parent *block             // block containing this label declaration
	lstmt  *nodes.LabeledStmt // statement declaring the label
	used   bool               // whether the label is used or not
}

type block struct {
	parent *block             // immediately enclosing block, or nil
	start  src.Pos            // start of block
	lstmt  *nodes.LabeledStmt // labeled statement associated with this block, or nil
}

func (ls *labelScope) err(pos src.Pos, format string, args ...interface{}) {
	ls.errh(Error{pos, fmt.Sprintf(format, args...)})
}

// declare declares the label introduced by s in block b and returns
// the new label. If the label was already declared, declare reports
// and error and the existing label is returned instead.
func (ls *labelScope) declare(b *block, s *nodes.LabeledStmt) *label {
	name := s.Label.Value
	labels := ls.labels
	if labels == nil {
		labels = make(map[string]*label)
		ls.labels = labels
	} else if alt := labels[name]; alt != nil {
		ls.err(s.Pos(), "label %s already defined at %s", name, alt.lstmt.Label.Pos().String())
		return alt
	}
	l := &label{b, s, false}
	labels[name] = l
	return l
}

// gotoTarget returns the labeled statement matching the given name and
// declared in block b or any of its enclosing blocks. The result is nil
// if the label is not defined, or doesn't match a valid labeled statement.
func (ls *labelScope) gotoTarget(b *block, name string) *nodes.LabeledStmt {
	if l := ls.labels[name]; l != nil {
		l.used = true // even if it's not a valid target
		for ; b != nil; b = b.parent {
			if l.parent == b {
				return l.lstmt
			}
		}
	}
	return nil
}

var invalid = new(nodes.LabeledStmt) // singleton to signal invalid enclosing target

// enclosingTarget returns the innermost enclosing labeled statement matching
// the given name. The result is nil if the label is not defined, and invalid
// if the label is defined but doesn't label a valid labeled statement.
func (ls *labelScope) enclosingTarget(b *block, name string) *nodes.LabeledStmt {
	if l := ls.labels[name]; l != nil {
		l.used = true // even if it's not a valid target (see e.g., test/fixedbugs/bug136.go)
		for ; b != nil; b = b.parent {
			if l.lstmt == b.lstmt {
				return l.lstmt
			}
		}
		return invalid
	}
	return nil
}

// targets describes the target statements within which break
// or continue statements are valid.
type targets struct {
	breaks    nodes.Stmt     // *ForStmt, *SwitchStmt, *SelectStmt, or nil
	continues *nodes.ForStmt // or nil
}

// blockBranches processes a block's body starting at start and returns the
// list of unresolved (forward) gotos. parent is the immediately enclosing
// block (or nil), ctxt provides information about the enclosing statements,
// and lstmt is the labeled statement associated with this block, or nil.
func (ls *labelScope) blockBranches(parent *block, ctxt targets, lstmt *nodes.LabeledStmt, start src.Pos, body []nodes.Stmt) []*nodes.BranchStmt {
	b := &block{parent: parent, start: start, lstmt: lstmt}

	var varPos src.Pos
	var varName nodes.Expr
	var fwdGotos, badGotos []*nodes.BranchStmt

	recordVarDecl := func(pos src.Pos, name nodes.Expr) {
		varPos = pos
		varName = name
		// Any existing forward goto jumping over the variable
		// declaration is invalid. The goto may still jump out
		// of the block and be ok, but we don't know that yet.
		// Remember all forward gotos as potential bad gotos.
		badGotos = append(badGotos[:0], fwdGotos...)
	}

	jumpsOverVarDecl := func(fwd *nodes.BranchStmt) bool {
		if varPos.IsKnown() {
			for _, bad := range badGotos {
				if fwd == bad {
					return true
				}
			}
		}
		return false
	}

	innerBlock := func(ctxt targets, start src.Pos, body []nodes.Stmt) {
		// Unresolved forward gotos from the inner block
		// become forward gotos for the current block.
		fwdGotos = append(fwdGotos, ls.blockBranches(b, ctxt, lstmt, start, body)...)
	}

	for _, stmt := range body {
		lstmt = nil
	L:
		switch s := stmt.(type) {
		case *nodes.DeclStmt:
			for _, d := range s.DeclList {
				if v, ok := d.(*nodes.VarDecl); ok {
					recordVarDecl(v.Pos(), v.NameList[0])
					break // the first VarDecl will do
				}
			}

		case *nodes.LabeledStmt:
			// declare non-blank label
			if name := s.Label.Value; name != "_" {
				l := ls.declare(b, s)
				// resolve matching forward gotos
				i := 0
				for _, fwd := range fwdGotos {
					if fwd.Label.Value == name {
						fwd.Target = s
						l.used = true
						if jumpsOverVarDecl(fwd) {
							ls.err(
								fwd.Label.Pos(),
								"goto %s jumps over declaration of %s at %s",
								name, String(varName), varPos,
							)
						}
					} else {
						// no match - keep forward goto
						fwdGotos[i] = fwd
						i++
					}
				}
				fwdGotos = fwdGotos[:i]
				lstmt = s
			}
			// process labeled statement
			stmt = s.Stmt
			goto L

		case *nodes.BranchStmt:
			// unlabeled branch statement
			if s.Label == nil {
				switch s.Tok {
				case nodes.BreakT:
					if t := ctxt.breaks; t != nil {
						s.Target = t
					} else {
						ls.err(s.Pos(), "break is not in a loop, switch, or select")
					}
				case nodes.ContinueT:
					if t := ctxt.continues; t != nil {
						s.Target = t
					} else {
						ls.err(s.Pos(), "continue is not in a loop")
					}
				case nodes.FallthroughT:
					// nothing to do
				case nodes.GotoT:
					fallthrough // should always have a label
				default:
					panic("invalid BranchStmt")
				}
				break
			}

			// labeled branch statement
			name := s.Label.Value
			switch s.Tok {
			case nodes.BreakT:
				// spec: "If there is a label, it must be that of an enclosing
				// "for", "switch", or "select" statement, and that is the one
				// whose execution terminates."
				if t := ls.enclosingTarget(b, name); t != nil {
					switch t := t.Stmt.(type) {
					case *nodes.SwitchStmt, *nodes.SelectStmt, *nodes.ForStmt:
						s.Target = t
					default:
						ls.err(s.Label.Pos(), "invalid break label %s", name)
					}
				} else {
					ls.err(s.Label.Pos(), "break label not defined: %s", name)
				}

			case nodes.ContinueT:
				// spec: "If there is a label, it must be that of an enclosing
				// "for" statement, and that is the one whose execution advances."
				if t := ls.enclosingTarget(b, name); t != nil {
					if t, ok := t.Stmt.(*nodes.ForStmt); ok {
						s.Target = t
					} else {
						ls.err(s.Label.Pos(), "invalid continue label %s", name)
					}
				} else {
					ls.err(s.Label.Pos(), "continue label not defined: %s", name)
				}

			case nodes.GotoT:
				if t := ls.gotoTarget(b, name); t != nil {
					s.Target = t
				} else {
					// label may be declared later - add goto to forward gotos
					fwdGotos = append(fwdGotos, s)
				}

			case nodes.FallthroughT:
				fallthrough // should never have a label
			default:
				panic("invalid BranchStmt")
			}

		case *nodes.AssignStmt:
			if s.Op == nodes.Def {
				recordVarDecl(s.Pos(), s.Lhs)
			}

		case *nodes.BlockStmt:
			innerBlock(ctxt, s.Pos(), s.List)

		case *nodes.IfStmt:
			innerBlock(ctxt, s.Then.Pos(), s.Then.List)
			if s.Else != nil {
				innerBlock(ctxt, s.Else.Pos(), []nodes.Stmt{s.Else})
			}

		case *nodes.ForStmt:
			innerBlock(targets{s, s}, s.Body.Pos(), s.Body.List)

		case *nodes.SwitchStmt:
			inner := targets{s, ctxt.continues}
			for _, cc := range s.Body {
				innerBlock(inner, cc.Pos(), cc.Body)
			}

		case *nodes.SelectStmt:
			inner := targets{s, ctxt.continues}
			for _, cc := range s.Body {
				innerBlock(inner, cc.Pos(), cc.Body)
			}
		}
	}

	return fwdGotos
}
