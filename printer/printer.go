// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package printer clones go/printer but with debug on.

Use like so:
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, f)
	fmt.Printf("%s", &buf)
*/
package printer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode"
	"unicode/utf8"
)

// ================================================================================
// This file implements printing of AST nodes; specifically
// expressions, statements, declarations, and files. It uses
// the print functionality implemented in printer.go.

const (
	maxNewlines = 2     // max. number of newlines between source text
	debug       = true  // enable for debugging
	infinity    = 1 << 30
)

type whiteSpace byte

const (
	ignore   = whiteSpace(0)
	blank    = whiteSpace(' ')
	vtab     = whiteSpace('\v')
	newline  = whiteSpace('\n')
	formfeed = whiteSpace('\f')
	indent   = whiteSpace('>')
	unindent = whiteSpace('<')
)

// A pmode value represents the current printer mode.
type pmode int

const (
	noExtraBlank     pmode = 1 << iota // disables extra blank after /*-style comment
	noExtraLinebreak                   // disables extra line break after /*-style comment
)

type commentInfo struct {
	cindex         int               // current comment index
	comment        *ast.CommentGroup // = printer.comments[cindex]; or nil
	commentOffset  int               // = printer.posFor(printer.comments[cindex].List[0].Pos()).Offset; or infinity
	commentNewline bool              // true if the comment group contains newlines
}

type printer struct {
	// Configuration (does not change after initialization)
	Config
	fset *token.FileSet

	// Current state
	output      []byte       // raw printer result
	indent      int          // current indentation
	mode        pmode        // current printer mode
	impliedSemi bool         // if set, a linebreak implies a semicolon
	lastTok     token.Token  // last token printed (token.ILLEGAL if it's whitespace)
	prevOpen    token.Token  // previous non-brace "open" token (, [, or token.ILLEGAL
	wsbuf       []whiteSpace // delayed white space

	// Positions
	// The out position differs from the pos position when the result
	// formatting differs from the source formatting (in the amount of
	// white space). If there's a difference and SourcePos is set in
	// ConfigMode, //line comments are used in the output to restore
	// original source positions for a reader.
	pos     token.Position // current position in AST (source) space
	out     token.Position // current position in output space
	last    token.Position // value of pos after calling writeString
	linePtr *int           // if set, record out.Line for the next token in *linePtr

	// The list of all source comments, in order of appearance.
	comments        []*ast.CommentGroup // may be nil
	useNodeComments bool                // if not set, ignore lead and line comments of nodes

	// Information about p.comments[p.cindex]; set up by nextComment.
	commentInfo

	// Cache of already computed node sizes.
	nodeSizes map[ast.Node]int

	// Cache of most recently computed line position.
	cachedPos  token.Pos
	cachedLine int // line corresponding to cachedPos
}

func (p *printer) init(cfg *Config, fset *token.FileSet, nodeSizes map[ast.Node]int) {
	p.Config = *cfg
	p.fset = fset
	p.pos = token.Position{Line: 1, Column: 1}
	p.out = token.Position{Line: 1, Column: 1}
	p.wsbuf = make([]whiteSpace, 0, 16) // whitespace sequences are short
	p.nodeSizes = nodeSizes
	p.cachedPos = -1
}

func (p *printer) internalError(msg ...interface{}) {
	if debug {
		fmt.Print(p.pos.String() + ": ")
		fmt.Println(msg...)
		panic("go/printer")
	}
}

// commentsHaveNewline reports whether a list of comments belonging to
// an *ast.CommentGroup contains newlines. Because the position information
// may only be partially correct, we also have to read the comment text.
func (p *printer) commentsHaveNewline(list []*ast.Comment) bool {
	// len(list) > 0
	line := p.lineFor(list[0].Pos())
	for i, c := range list {
		if i > 0 && p.lineFor(list[i].Pos()) != line {
			// not all comments on the same line
			return true
		}
		if t := c.Text; len(t) >= 2 && (t[1] == '/' || strings.Contains(t, "\n")) {
			return true
		}
	}
	_ = line
	return false
}

func (p *printer) nextComment() {
	for p.cindex < len(p.comments) {
		c := p.comments[p.cindex]
		p.cindex++
		if list := c.List; len(list) > 0 {
			p.comment = c
			p.commentOffset = p.posFor(list[0].Pos()).Offset
			p.commentNewline = p.commentsHaveNewline(list)
			return
		}
		// we should not reach here (correct ASTs don't have empty
		// ast.CommentGroup nodes), but be conservative and try again
	}
	// no more comments
	p.commentOffset = infinity
}

// commentBefore reports whether the current comment group occurs
// before the next position in the source code and printing it does
// not introduce implicit semicolons.
//
func (p *printer) commentBefore(next token.Position) bool {
	return p.commentOffset < next.Offset && (!p.impliedSemi || !p.commentNewline)
}

// commentSizeBefore returns the estimated size of the
// comments on the same line before the next position.
//
func (p *printer) commentSizeBefore(next token.Position) int {
	// save/restore current p.commentInfo (p.nextComment() modifies it)
	defer func(info commentInfo) {
		p.commentInfo = info
	}(p.commentInfo)

	size := 0
	for p.commentBefore(next) {
		for _, c := range p.comment.List {
			size += len(c.Text)
		}
		p.nextComment()
	}
	return size
}

// recordLine records the output line number for the next non-whitespace
// token in *linePtr. It is used to compute an accurate line number for a
// formatted construct, independent of pending (not yet emitted) whitespace
// or comments.
//
func (p *printer) recordLine(linePtr *int) {
	p.linePtr = linePtr
}

// linesFrom returns the number of output lines between the current
// output line and the line argument, ignoring any pending (not yet
// emitted) whitespace or comments. It is used to compute an accurate
// size (in number of lines) for a formatted construct.
//
func (p *printer) linesFrom(line int) int {
	return p.out.Line - line
}

func (p *printer) posFor(pos token.Pos) token.Position {
	// not used frequently enough to cache entire token.Position
	return p.fset.Position(pos)
}

func (p *printer) lineFor(pos token.Pos) int {
	if pos != p.cachedPos {
		p.cachedPos = pos
		p.cachedLine = p.fset.Position(pos).Line
	}
	return p.cachedLine
}

// atLineBegin emits a //line comment if necessary and prints indentation.
func (p *printer) atLineBegin(pos token.Position) {
	// write a //line comment if necessary
	if p.Config.Mode&SourcePos != 0 && pos.IsValid() && (p.out.Line != pos.Line || p.out.Filename != pos.Filename) {
		p.output = append(p.output, tabwriter.Escape) // protect '\n' in //line from tabwriter interpretation
		p.output = append(p.output, fmt.Sprintf("//line %s:%d\n", pos.Filename, pos.Line)...)
		p.output = append(p.output, tabwriter.Escape)
		// p.out must match the //line comment
		p.out.Filename = pos.Filename
		p.out.Line = pos.Line
	}

	// write indentation
	// use "hard" htabs - indentation columns
	// must not be discarded by the tabwriter
	n := p.Config.Indent + p.indent // include base indentation
	for i := 0; i < n; i++ {
		p.output = append(p.output, '\t')
	}

	// update positions
	p.pos.Offset += n
	p.pos.Column += n
	p.out.Column += n
}

// writeByte writes ch n times to p.output and updates p.pos.
func (p *printer) writeByte(ch byte, n int) {
	if p.out.Column == 1 {
		p.atLineBegin(p.pos)
	}

	for i := 0; i < n; i++ {
		p.output = append(p.output, ch)
	}

	// update positions
	p.pos.Offset += n
	if ch == '\n' || ch == '\f' {
		p.pos.Line += n
		p.out.Line += n
		p.pos.Column = 1
		p.out.Column = 1
		return
	}
	p.pos.Column += n
	p.out.Column += n
}

// writeString writes the string s to p.output and updates p.pos, p.out,
// and p.last. If isLit is set, s is escaped w/ tabwriter.Escape characters
// to protect s from being interpreted by the tabwriter.
//
// Note: writeString is only used to write Go tokens, literals, and
// comments, all of which must be written literally. Thus, it is correct
// to always set isLit = true. However, setting it explicitly only when
// needed (i.e., when we don't know that s contains no tabs or line breaks)
// avoids processing extra escape characters and reduces run time of the
// printer benchmark by up to 10%.
//
func (p *printer) writeString(pos token.Position, s string, isLit bool) {
	if p.out.Column == 1 {
		p.atLineBegin(pos)
	}

	if pos.IsValid() {
		// update p.pos (if pos is invalid, continue with existing p.pos)
		// Note: Must do this after handling line beginnings because
		// atLineBegin updates p.pos if there's indentation, but p.pos
		// is the position of s.
		p.pos = pos
	}

	if isLit {
		// Protect s such that is passes through the tabwriter
		// unchanged. Note that valid Go programs cannot contain
		// tabwriter.Escape bytes since they do not appear in legal
		// UTF-8 sequences.
		p.output = append(p.output, tabwriter.Escape)
	}

	if debug {
		p.output = append(p.output, fmt.Sprintf("/*%s*/", pos)...) // do not update p.pos!
	}
	p.output = append(p.output, s...)

	// update positions
	nlines := 0
	var li int // index of last newline; valid if nlines > 0
	for i := 0; i < len(s); i++ {
		// Go tokens cannot contain '\f' - no need to look for it
		if s[i] == '\n' {
			nlines++
			li = i
		}
	}
	p.pos.Offset += len(s)
	if nlines > 0 {
		p.pos.Line += nlines
		p.out.Line += nlines
		c := len(s) - li
		p.pos.Column = c
		p.out.Column = c
	} else {
		p.pos.Column += len(s)
		p.out.Column += len(s)
	}

	if isLit {
		p.output = append(p.output, tabwriter.Escape)
	}

	p.last = p.pos
}

// writeCommentPrefix writes the whitespace before a comment.
// If there is any pending whitespace, it consumes as much of
// it as is likely to help position the comment nicely.
// pos is the comment position, next the position of the item
// after all pending comments, prev is the previous comment in
// a group of comments (or nil), and tok is the next token.
//
func (p *printer) writeCommentPrefix(pos, next token.Position, prev, comment *ast.Comment, tok token.Token) {
	if len(p.output) == 0 {
		// the comment is the first item to be printed - don't write any whitespace
		return
	}

	if pos.IsValid() && pos.Filename != p.last.Filename {
		// comment in a different file - separate with newlines
		p.writeByte('\f', maxNewlines)
		return
	}

	if pos.Line == p.last.Line && (prev == nil || prev.Text[1] != '/') {
		// comment on the same line as last item:
		// separate with at least one separator
		hasSep := false
		if prev == nil {
			// first comment of a comment group
			j := 0
			for i, ch := range p.wsbuf {
				switch ch {
				case blank:
					// ignore any blanks before a comment
					p.wsbuf[i] = ignore
					continue
				case vtab:
					// respect existing tabs - important
					// for proper formatting of commented structs
					hasSep = true
					continue
				case indent:
					// apply pending indentation
					continue
				}
				j = i
				break
			}
			p.writeWhitespace(j)
		}
		// make sure there is at least one separator
		if !hasSep {
			sep := byte('\t')
			if pos.Line == next.Line {
				// next item is on the same line as the comment
				// (which must be a /*-style comment): separate
				// with a blank instead of a tab
				sep = ' '
			}
			p.writeByte(sep, 1)
		}

	} else {
		// comment on a different line:
		// separate with at least one line break
		droppedLinebreak := false
		j := 0
		for i, ch := range p.wsbuf {
			switch ch {
			case blank, vtab:
				// ignore any horizontal whitespace before line breaks
				p.wsbuf[i] = ignore
				continue
			case indent:
				// apply pending indentation
				continue
			case unindent:
				// if this is not the last unindent, apply it
				// as it is (likely) belonging to the last
				// construct (e.g., a multi-line expression list)
				// and is not part of closing a block
				if i+1 < len(p.wsbuf) && p.wsbuf[i+1] == unindent {
					continue
				}
				// if the next token is not a closing }, apply the unindent
				// if it appears that the comment is aligned with the
				// token; otherwise assume the unindent is part of a
				// closing block and stop (this scenario appears with
				// comments before a case label where the comments
				// apply to the next case instead of the current one)
				if tok != token.RBRACE && pos.Column == next.Column {
					continue
				}
			case newline, formfeed:
				p.wsbuf[i] = ignore
				droppedLinebreak = prev == nil // record only if first comment of a group
			}
			j = i
			break
		}
		p.writeWhitespace(j)

		// determine number of linebreaks before the comment
		n := 0
		if pos.IsValid() && p.last.IsValid() {
			n = pos.Line - p.last.Line
			if n < 0 { // should never happen
				n = 0
			}
		}

		// at the package scope level only (p.indent == 0),
		// add an extra newline if we dropped one before:
		// this preserves a blank line before documentation
		// comments at the package scope level (issue 2570)
		if p.indent == 0 && droppedLinebreak {
			n++
		}

		// make sure there is at least one line break
		// if the previous comment was a line comment
		if n == 0 && prev != nil && prev.Text[1] == '/' {
			n = 1
		}

		if n > 0 {
			// use formfeeds to break columns before a comment;
			// this is analogous to using formfeeds to separate
			// individual lines of /*-style comments
			p.writeByte('\f', nlimit(n))
		}
	}
}

// Returns true if s contains only white space
// (only tabs and blanks can appear in the printer's context).
//
func isBlank(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > ' ' {
			return false
		}
	}
	return true
}

// commonPrefix returns the common prefix of a and b.
func commonPrefix(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] && (a[i] <= ' ' || a[i] == '*') {
		i++
	}
	return a[0:i]
}

// trimRight returns s with trailing whitespace removed.
func trimRight(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// stripCommonPrefix removes a common prefix from /*-style comment lines (unless no
// comment line is indented, all but the first line have some form of space prefix).
// The prefix is computed using heuristics such that is likely that the comment
// contents are nicely laid out after re-printing each line using the printer's
// current indentation.
//
func stripCommonPrefix(lines []string) {
	if len(lines) <= 1 {
		return // at most one line - nothing to do
	}
	// len(lines) > 1

	// The heuristic in this function tries to handle a few
	// common patterns of /*-style comments: Comments where
	// the opening /* and closing */ are aligned and the
	// rest of the comment text is aligned and indented with
	// blanks or tabs, cases with a vertical "line of stars"
	// on the left, and cases where the closing */ is on the
	// same line as the last comment text.

	// Compute maximum common white prefix of all but the first,
	// last, and blank lines, and replace blank lines with empty
	// lines (the first line starts with /* and has no prefix).
	// In cases where only the first and last lines are not blank,
	// such as two-line comments, or comments where all inner lines
	// are blank, consider the last line for the prefix computation
	// since otherwise the prefix would be empty.
	//
	// Note that the first and last line are never empty (they
	// contain the opening /* and closing */ respectively) and
	// thus they can be ignored by the blank line check.
	prefix := ""
	prefixSet := false
	if len(lines) > 2 {
		for i, line := range lines[1 : len(lines)-1] {
			if isBlank(line) {
				lines[1+i] = "" // range starts with lines[1]
			} else {
				if !prefixSet {
					prefix = line
					prefixSet = true
				}
				prefix = commonPrefix(prefix, line)
			}

		}
	}
	// If we don't have a prefix yet, consider the last line.
	if !prefixSet {
		line := lines[len(lines)-1]
		prefix = commonPrefix(line, line)
	}

	/*
	 * Check for vertical "line of stars" and correct prefix accordingly.
	 */
	lineOfStars := false
	if i := strings.Index(prefix, "*"); i >= 0 {
		// Line of stars present.
		if i > 0 && prefix[i-1] == ' ' {
			i-- // remove trailing blank from prefix so stars remain aligned
		}
		prefix = prefix[0:i]
		lineOfStars = true
	} else {
		// No line of stars present.
		// Determine the white space on the first line after the /*
		// and before the beginning of the comment text, assume two
		// blanks instead of the /* unless the first character after
		// the /* is a tab. If the first comment line is empty but
		// for the opening /*, assume up to 3 blanks or a tab. This
		// whitespace may be found as suffix in the common prefix.
		first := lines[0]
		if isBlank(first[2:]) {
			// no comment text on the first line:
			// reduce prefix by up to 3 blanks or a tab
			// if present - this keeps comment text indented
			// relative to the /* and */'s if it was indented
			// in the first place
			i := len(prefix)
			for n := 0; n < 3 && i > 0 && prefix[i-1] == ' '; n++ {
				i--
			}
			if i == len(prefix) && i > 0 && prefix[i-1] == '\t' {
				i--
			}
			prefix = prefix[0:i]
		} else {
			// comment text on the first line
			suffix := make([]byte, len(first))
			n := 2 // start after opening /*
			for n < len(first) && first[n] <= ' ' {
				suffix[n] = first[n]
				n++
			}
			if n > 2 && suffix[2] == '\t' {
				// assume the '\t' compensates for the /*
				suffix = suffix[2:n]
			} else {
				// otherwise assume two blanks
				suffix[0], suffix[1] = ' ', ' '
				suffix = suffix[0:n]
			}
			// Shorten the computed common prefix by the length of
			// suffix, if it is found as suffix of the prefix.
			prefix = strings.TrimSuffix(prefix, string(suffix))
		}
	}

	// Handle last line: If it only contains a closing */, align it
	// with the opening /*, otherwise align the text with the other
	// lines.
	last := lines[len(lines)-1]
	closing := "*/"
	i := strings.Index(last, closing) // i >= 0 (closing is always present)
	if isBlank(last[0:i]) {
		// last line only contains closing */
		if lineOfStars {
			closing = " */" // add blank to align final star
		}
		lines[len(lines)-1] = prefix + closing
	} else {
		// last line contains more comment text - assume
		// it is aligned like the other lines and include
		// in prefix computation
		prefix = commonPrefix(prefix, last)
	}

	// Remove the common prefix from all but the first and empty lines.
	for i, line := range lines {
		if i > 0 && line != "" {
			lines[i] = line[len(prefix):]
		}
	}
}

func (p *printer) writeComment(comment *ast.Comment) {
	text := comment.Text
	pos := p.posFor(comment.Pos())

	const linePrefix = "//line "
	if strings.HasPrefix(text, linePrefix) && (!pos.IsValid() || pos.Column == 1) {
		// possibly a line directive
		ldir := strings.TrimSpace(text[len(linePrefix):])
		if i := strings.LastIndex(ldir, ":"); i >= 0 {
			if line, err := strconv.Atoi(ldir[i+1:]); err == nil && line > 0 {
				// The line directive we are about to print changed
				// the Filename and Line number used for subsequent
				// tokens. We have to update our AST-space position
				// accordingly and suspend indentation temporarily.
				indent := p.indent
				p.indent = 0
				defer func() {
					p.pos.Filename = ldir[:i]
					p.pos.Line = line
					p.pos.Column = 1
					p.indent = indent
				}()
			}
		}
	}

	// shortcut common case of //-style comments
	if text[1] == '/' {
		p.writeString(pos, trimRight(text), true)
		return
	}

	// for /*-style comments, print line by line and let the
	// write function take care of the proper indentation
	lines := strings.Split(text, "\n")

	// The comment started in the first column but is going
	// to be indented. For an idempotent result, add indentation
	// to all lines such that they look like they were indented
	// before - this will make sure the common prefix computation
	// is the same independent of how many times formatting is
	// applied (was issue 1835).
	if pos.IsValid() && pos.Column == 1 && p.indent > 0 {
		for i, line := range lines[1:] {
			lines[1+i] = "   " + line
		}
	}

	stripCommonPrefix(lines)

	// write comment lines, separated by formfeed,
	// without a line break after the last line
	for i, line := range lines {
		if i > 0 {
			p.writeByte('\f', 1)
			pos = p.pos
		}
		if len(line) > 0 {
			p.writeString(pos, trimRight(line), true)
		}
	}
}

// writeCommentSuffix writes a line break after a comment if indicated
// and processes any leftover indentation information. If a line break
// is needed, the kind of break (newline vs formfeed) depends on the
// pending whitespace. The writeCommentSuffix result indicates if a
// newline was written or if a formfeed was dropped from the whitespace
// buffer.
//
func (p *printer) writeCommentSuffix(needsLinebreak bool) (wroteNewline, droppedFF bool) {
	for i, ch := range p.wsbuf {
		switch ch {
		case blank, vtab:
			// ignore trailing whitespace
			p.wsbuf[i] = ignore
		case indent, unindent:
			// don't lose indentation information
		case newline, formfeed:
			// if we need a line break, keep exactly one
			// but remember if we dropped any formfeeds
			if needsLinebreak {
				needsLinebreak = false
				wroteNewline = true
			} else {
				if ch == formfeed {
					droppedFF = true
				}
				p.wsbuf[i] = ignore
			}
		}
	}
	p.writeWhitespace(len(p.wsbuf))

	// make sure we have a line break
	if needsLinebreak {
		p.writeByte('\n', 1)
		wroteNewline = true
	}

	return
}

// intersperseComments consumes all comments that appear before the next token
// tok and prints it together with the buffered whitespace (i.e., the whitespace
// that needs to be written before the next token). A heuristic is used to mix
// the comments and whitespace. The intersperseComments result indicates if a
// newline was written or if a formfeed was dropped from the whitespace buffer.
//
func (p *printer) intersperseComments(next token.Position, tok token.Token) (wroteNewline, droppedFF bool) {
	var last *ast.Comment
	for p.commentBefore(next) {
		for _, c := range p.comment.List {
			p.writeCommentPrefix(p.posFor(c.Pos()), next, last, c, tok)
			p.writeComment(c)
			last = c
		}
		p.nextComment()
	}

	if last != nil {
		// if the last comment is a /*-style comment and the next item
		// follows on the same line but is not a comma, and not a "closing"
		// token immediately following its corresponding "opening" token,
		// add an extra blank for separation unless explicitly disabled
		if p.mode&noExtraBlank == 0 &&
			last.Text[1] == '*' && p.lineFor(last.Pos()) == next.Line &&
			tok != token.COMMA &&
			(tok != token.RPAREN || p.prevOpen == token.LPAREN) &&
			(tok != token.RBRACK || p.prevOpen == token.LBRACK) {
			p.writeByte(' ', 1)
		}
		// ensure that there is a line break after a //-style comment,
		// before a closing '}' unless explicitly disabled, or at eof
		needsLinebreak :=
			last.Text[1] == '/' ||
				tok == token.RBRACE && p.mode&noExtraLinebreak == 0 ||
				tok == token.EOF
		return p.writeCommentSuffix(needsLinebreak)
	}

	// no comment was written - we should never reach here since
	// intersperseComments should not be called in that case
	p.internalError("intersperseComments called without pending comments")
	return
}

// whiteWhitespace writes the first n whitespace entries.
func (p *printer) writeWhitespace(n int) {
	// write entries
	for i := 0; i < n; i++ {
		switch ch := p.wsbuf[i]; ch {
		case ignore:
			// ignore!
		case indent:
			p.indent++
		case unindent:
			p.indent--
			if p.indent < 0 {
				p.internalError("negative indentation:", p.indent)
				p.indent = 0
			}
		case newline, formfeed:
			// A line break immediately followed by a "correcting"
			// unindent is swapped with the unindent - this permits
			// proper label positioning. If a comment is between
			// the line break and the label, the unindent is not
			// part of the comment whitespace prefix and the comment
			// will be positioned correctly indented.
			if i+1 < n && p.wsbuf[i+1] == unindent {
				// Use a formfeed to terminate the current section.
				// Otherwise, a long label name on the next line leading
				// to a wide column may increase the indentation column
				// of lines before the label; effectively leading to wrong
				// indentation.
				p.wsbuf[i], p.wsbuf[i+1] = unindent, formfeed
				i-- // do it again
				continue
			}
			fallthrough
		default:
			p.writeByte(byte(ch), 1)
		}
	}

	// shift remaining entries down
	l := copy(p.wsbuf, p.wsbuf[n:])
	p.wsbuf = p.wsbuf[:l]
}

// ----------------------------------------------------------------------------
// Printing interface

// nlines limits n to maxNewlines.
func nlimit(n int) int {
	if n > maxNewlines {
		n = maxNewlines
	}
	return n
}

func mayCombine(prev token.Token, next byte) (b bool) {
	switch prev {
	case token.INT:
		b = next == '.' // 1.
	case token.ADD:
		b = next == '+' // ++
	case token.SUB:
		b = next == '-' // --
	case token.QUO:
		b = next == '*' // /*
	case token.LSS:
		b = next == '-' || next == '<' // <- or <<
	case token.AND:
		b = next == '&' || next == '^' // && or &^
	}
	return
}

// print prints a list of "items" (roughly corresponding to syntactic
// tokens, but also including whitespace and formatting information).
// It is the only print function that should be called directly from
// any of the AST printing functions in nodes.go.
//
// Whitespace is accumulated until a non-whitespace token appears. Any
// comments that need to appear before that token are printed first,
// taking into account the amount and structure of any pending white-
// space for best comment placement. Then, any leftover whitespace is
// printed, followed by the actual token.
//
func (p *printer) print(args ...interface{}) {
	for _, arg := range args {
		// information about the current arg
		var data string
		var isLit bool
		var impliedSemi bool // value for p.impliedSemi after this arg

		// record previous opening token, if any
		switch p.lastTok {
		case token.ILLEGAL:
			// ignore (white space)
		case token.LPAREN, token.LBRACK:
			p.prevOpen = p.lastTok
		default:
			// other tokens followed any opening token
			p.prevOpen = token.ILLEGAL
		}

		switch x := arg.(type) {
		case pmode:
			// toggle printer mode
			p.mode ^= x
			continue

		case whiteSpace:
			if x == ignore {
				// don't add ignore's to the buffer; they
				// may screw up "correcting" unindents (see
				// LabeledStmt)
				continue
			}
			i := len(p.wsbuf)
			if i == cap(p.wsbuf) {
				// Whitespace sequences are very short so this should
				// never happen. Handle gracefully (but possibly with
				// bad comment placement) if it does happen.
				p.writeWhitespace(i)
				i = 0
			}
			p.wsbuf = p.wsbuf[0 : i+1]
			p.wsbuf[i] = x
			if x == newline || x == formfeed {
				// newlines affect the current state (p.impliedSemi)
				// and not the state after printing arg (impliedSemi)
				// because comments can be interspersed before the arg
				// in this case
				p.impliedSemi = false
			}
			p.lastTok = token.ILLEGAL
			continue

		case *ast.Ident:
			data = x.Name
			impliedSemi = true
			p.lastTok = token.IDENT

		case *ast.BasicLit:
			data = x.Value
			isLit = true
			impliedSemi = true
			p.lastTok = x.Kind

		case token.Token:
			s := x.String()
			if mayCombine(p.lastTok, s[0]) {
				// the previous and the current token must be
				// separated by a blank otherwise they combine
				// into a different incorrect token sequence
				// (except for token.INT followed by a '.' this
				// should never happen because it is taken care
				// of via binary expression formatting)
				if len(p.wsbuf) != 0 {
					p.internalError("whitespace buffer not empty")
				}
				p.wsbuf = p.wsbuf[0:1]
				p.wsbuf[0] = ' '
			}
			data = s
			// some keywords followed by a newline imply a semicolon
			switch x {
			case token.BREAK, token.CONTINUE, token.FALLTHROUGH, token.RETURN,
				token.INC, token.DEC, token.RPAREN, token.RBRACK, token.RBRACE:
				impliedSemi = true
			}
			p.lastTok = x

		case token.Pos:
			if x.IsValid() {
				p.pos = p.posFor(x) // accurate position of next item
			}
			continue

		case string:
			// incorrect AST - print error message
			data = x
			isLit = true
			impliedSemi = true
			p.lastTok = token.STRING

		default:
			fmt.Fprintf(os.Stderr, "print: unsupported argument %v (%T)\n", arg, arg)
			panic("go/printer type")
		}
		// data != ""

		next := p.pos // estimated/accurate position of next item
		wroteNewline, droppedFF := p.flush(next, p.lastTok)

		// intersperse extra newlines if present in the source and
		// if they don't cause extra semicolons (don't do this in
		// flush as it will cause extra newlines at the end of a file)
		if !p.impliedSemi {
			n := nlimit(next.Line - p.pos.Line)
			// don't exceed maxNewlines if we already wrote one
			if wroteNewline && n == maxNewlines {
				n = maxNewlines - 1
			}
			if n > 0 {
				ch := byte('\n')
				if droppedFF {
					ch = '\f' // use formfeed since we dropped one before
				}
				p.writeByte(ch, n)
				impliedSemi = false
			}
		}

		// the next token starts now - record its line number if requested
		if p.linePtr != nil {
			*p.linePtr = p.out.Line
			p.linePtr = nil
		}

		p.writeString(next, data, isLit)
		p.impliedSemi = impliedSemi
	}
}

// flush prints any pending comments and whitespace occurring textually
// before the position of the next token tok. The flush result indicates
// if a newline was written or if a formfeed was dropped from the whitespace
// buffer.
//
func (p *printer) flush(next token.Position, tok token.Token) (wroteNewline, droppedFF bool) {
	if p.commentBefore(next) {
		// if there are comments before the next item, intersperse them
		wroteNewline, droppedFF = p.intersperseComments(next, tok)
	} else {
		// otherwise, write any leftover whitespace
		p.writeWhitespace(len(p.wsbuf))
	}
	return
}

// getNode returns the ast.CommentGroup associated with n, if any.
func getDoc(n ast.Node) *ast.CommentGroup {
	switch n := n.(type) {
	case *ast.Field:
		return n.Doc
	case *ast.ImportSpec:
		return n.Doc
	case *ast.ValueSpec:
		return n.Doc
	case *ast.TypeSpec:
		return n.Doc
	case *ast.GenDecl:
		return n.Doc
	case *ast.FuncDecl:
		return n.Doc
	case *ast.File:
		return n.Doc
	}
	return nil
}

func (p *printer) printNode(node interface{}) error {
	// unpack *CommentedNode, if any
	var comments []*ast.CommentGroup
	if cnode, ok := node.(*CommentedNode); ok {
		node = cnode.Node
		comments = cnode.Comments
	}

	if comments != nil {
		// commented node - restrict comment list to relevant range
		n, ok := node.(ast.Node)
		if !ok {
			goto unsupported
		}
		beg := n.Pos()
		end := n.End()
		// if the node has associated documentation,
		// include that commentgroup in the range
		// (the comment list is sorted in the order
		// of the comment appearance in the source code)
		if doc := getDoc(n); doc != nil {
			beg = doc.Pos()
		}
		// token.Pos values are global offsets, we can
		// compare them directly
		i := 0
		for i < len(comments) && comments[i].End() < beg {
			i++
		}
		j := i
		for j < len(comments) && comments[j].Pos() < end {
			j++
		}
		if i < j {
			p.comments = comments[i:j]
		}
	} else if n, ok := node.(*ast.File); ok {
		// use ast.File comments, if any
		p.comments = n.Comments
	}

	// if there are no comments, use node comments
	p.useNodeComments = p.comments == nil

	// get comments ready for use
	p.nextComment()

	// format node
	switch n := node.(type) {
	case ast.Expr:
		p.expr(n)
	case ast.Stmt:
		// A labeled statement will un-indent to position the label.
		// Set p.indent to 1 so we don't get indent "underflow".
		if _, ok := n.(*ast.LabeledStmt); ok {
			p.indent = 1
		}
		p.stmt(n, false)
	case ast.Decl:
		p.decl(n)
	case ast.Spec:
		p.spec(n, 1, false)
	case []ast.Stmt:
		// A labeled statement will un-indent to position the label.
		// Set p.indent to 1 so we don't get indent "underflow".
		for _, s := range n {
			if _, ok := s.(*ast.LabeledStmt); ok {
				p.indent = 1
			}
		}
		p.stmtList(n, 0, false)
	case []ast.Decl:
		p.declList(n)
	case *ast.File:
		p.file(n)
	default:
		goto unsupported
	}

	return nil

unsupported:
	return fmt.Errorf("go/printer: unsupported node type %T", node)
}

// ----------------------------------------------------------------------------
// Trimmer

// A trimmer is an io.Writer filter for stripping tabwriter.Escape
// characters, trailing blanks and tabs, and for converting formfeed
// and vtab characters into newlines and htabs (in case no tabwriter
// is used). Text bracketed by tabwriter.Escape characters is passed
// through unchanged.
//
type trimmer struct {
	output io.Writer
	state  int
	space  []byte
}

// trimmer is implemented as a state machine.
// It can be in one of the following states:
const (
	inSpace  = iota // inside space
	inEscape        // inside text bracketed by tabwriter.Escapes
	inText          // inside text
)

func (p *trimmer) resetSpace() {
	p.state = inSpace
	p.space = p.space[0:0]
}

// Design note: It is tempting to eliminate extra blanks occurring in
//              whitespace in this function as it could simplify some
//              of the blanks logic in the node printing functions.
//              However, this would mess up any formatting done by
//              the tabwriter.

var aNewline = []byte("\n")

func (p *trimmer) Write(data []byte) (n int, err error) {
	// invariants:
	// p.state == inSpace:
	//	p.space is unwritten
	// p.state == inEscape, inText:
	//	data[m:n] is unwritten
	m := 0
	var b byte
	for n, b = range data {
		if b == '\v' {
			b = '\t' // convert to htab
		}
		switch p.state {
		case inSpace:
			switch b {
			case '\t', ' ':
				p.space = append(p.space, b)
			case '\n', '\f':
				p.resetSpace() // discard trailing space
				_, err = p.output.Write(aNewline)
			case tabwriter.Escape:
				_, err = p.output.Write(p.space)
				p.state = inEscape
				m = n + 1 // +1: skip tabwriter.Escape
			default:
				_, err = p.output.Write(p.space)
				p.state = inText
				m = n
			}
		case inEscape:
			if b == tabwriter.Escape {
				_, err = p.output.Write(data[m:n])
				p.resetSpace()
			}
		case inText:
			switch b {
			case '\t', ' ':
				_, err = p.output.Write(data[m:n])
				p.resetSpace()
				p.space = append(p.space, b)
			case '\n', '\f':
				_, err = p.output.Write(data[m:n])
				p.resetSpace()
				if err == nil {
					_, err = p.output.Write(aNewline)
				}
			case tabwriter.Escape:
				_, err = p.output.Write(data[m:n])
				p.state = inEscape
				m = n + 1 // +1: skip tabwriter.Escape
			}
		default:
			panic("unreachable")
		}
		if err != nil {
			return
		}
	}
	n = len(data)

	switch p.state {
	case inEscape, inText:
		_, err = p.output.Write(data[m:n])
		p.resetSpace()
	}

	return
}

// ----------------------------------------------------------------------------
// Public interface

// A Mode value is a set of flags (or 0). They control printing.
type Mode uint

const (
	RawFormat Mode = 1 << iota // do not use a tabwriter; if set, UseSpaces is ignored
	TabIndent                  // use tabs for indentation independent of UseSpaces
	UseSpaces                  // use spaces instead of tabs for alignment
	SourcePos                  // emit //line comments to preserve original source positions
)

// A Config node controls the output of Fprint.
type Config struct {
	Mode     Mode // default: 0
	Tabwidth int  // default: 8
	Indent   int  // default: 0 (all code is indented at least by this much)
}

// fprint implements Fprint and takes a nodesSizes map for setting up the printer state.
func (cfg *Config) fprint(output io.Writer, fset *token.FileSet, node interface{}, nodeSizes map[ast.Node]int) (err error) {
	// print node
	var p printer
	p.init(cfg, fset, nodeSizes)
	if err = p.printNode(node); err != nil {
		return
	}
	// print outstanding comments
	p.impliedSemi = false // EOF acts like a newline
	p.flush(token.Position{Offset: infinity, Line: infinity}, token.EOF)

	// redirect output through a trimmer to eliminate trailing whitespace
	// (Input to a tabwriter must be untrimmed since trailing tabs provide
	// formatting information. The tabwriter could provide trimming
	// functionality but no tabwriter is used when RawFormat is set.)
	output = &trimmer{output: output}

	// redirect output through a tabwriter if necessary
	if cfg.Mode&RawFormat == 0 {
		minwidth := cfg.Tabwidth

		padchar := byte('\t')
		if cfg.Mode&UseSpaces != 0 {
			padchar = ' '
		}

		twmode := tabwriter.DiscardEmptyColumns
		if cfg.Mode&TabIndent != 0 {
			minwidth = 0
			twmode |= tabwriter.TabIndent
		}

		output = tabwriter.NewWriter(output, minwidth, cfg.Tabwidth, 1, padchar, twmode)
	}

	// write printer result via tabwriter/trimmer to output
	if _, err = output.Write(p.output); err != nil {
		return
	}

	// flush tabwriter, if any
	if tw, _ := output.(*tabwriter.Writer); tw != nil {
		err = tw.Flush()
	}

	return
}

// A CommentedNode bundles an AST node and corresponding comments.
// It may be provided as argument to any of the Fprint functions.
//
type CommentedNode struct {
	Node     interface{} // *ast.File, or ast.Expr, ast.Decl, ast.Spec, or ast.Stmt
	Comments []*ast.CommentGroup
}

// Fprint "pretty-prints" an AST node to output for a given configuration cfg.
// Position information is interpreted relative to the file set fset.
// The node type must be *ast.File, *CommentedNode, []ast.Decl, []ast.Stmt,
// or assignment-compatible to ast.Expr, ast.Decl, ast.Spec, or ast.Stmt.
//
func (cfg *Config) Fprint(output io.Writer, fset *token.FileSet, node interface{}) error {
	return cfg.fprint(output, fset, node, make(map[ast.Node]int))
}

// Fprint "pretty-prints" an AST node to output.
// It calls Config.Fprint with default settings.
//
func Fprint(output io.Writer, fset *token.FileSet, node interface{}) error {
	return (&Config{Tabwidth: 8}).Fprint(output, fset, node)
}


// ================================================================================

// Formatting issues:
// - better comment formatting for /*-style comments at the end of a line (e.g. a declaration)
//   when the comment spans multiple lines; if such a comment is just two lines, formatting is
//   not idempotent
// - formatting of expression lists
// - should use blank instead of tab to separate one-line function bodies from
//   the function header unless there is a group of consecutive one-liners

// ----------------------------------------------------------------------------
// Common AST nodes.

// Print as many newlines as necessary (but at least min newlines) to get to
// the current line. ws is printed before the first line break. If newSection
// is set, the first line break is printed as formfeed. Returns true if any
// line break was printed; returns false otherwise.
//
// TODO(gri): linebreak may add too many lines if the next statement at "line"
//            is preceded by comments because the computation of n assumes
//            the current position before the comment and the target position
//            after the comment. Thus, after interspersing such comments, the
//            space taken up by them is not considered to reduce the number of
//            linebreaks. At the moment there is no easy way to know about
//            future (not yet interspersed) comments in this function.
//
func (p *printer) linebreak(line, min int, ws whiteSpace, newSection bool) (printedBreak bool) {
	n := nlimit(line - p.pos.Line)
	if n < min {
		n = min
	}
	if n > 0 {
		p.print(ws)
		if newSection {
			p.print(formfeed)
			n--
		}
		for ; n > 0; n-- {
			p.print(newline)
		}
		printedBreak = true
	}
	return
}

// setComment sets g as the next comment if g != nil and if node comments
// are enabled - this mode is used when printing source code fragments such
// as exports only. It assumes that there is no pending comment in p.comments
// and at most one pending comment in the p.comment cache.
func (p *printer) setComment(g *ast.CommentGroup) {
	if g == nil || !p.useNodeComments {
		return
	}
	if p.comments == nil {
		// initialize p.comments lazily
		p.comments = make([]*ast.CommentGroup, 1)
	} else if p.cindex < len(p.comments) {
		// for some reason there are pending comments; this
		// should never happen - handle gracefully and flush
		// all comments up to g, ignore anything after that
		p.flush(p.posFor(g.List[0].Pos()), token.ILLEGAL)
		p.comments = p.comments[0:1]
		// in debug mode, report error
		p.internalError("setComment found pending comments")
	}
	p.comments[0] = g
	p.cindex = 0
	// don't overwrite any pending comment in the p.comment cache
	// (there may be a pending comment when a line comment is
	// immediately followed by a lead comment with no other
	// tokens between)
	if p.commentOffset == infinity {
		p.nextComment() // get comment ready for use
	}
}

type exprListMode uint

const (
	commaTerm exprListMode = 1 << iota // list is optionally terminated by a comma
	noIndent                           // no extra indentation in multi-line lists
)

// If indent is set, a multi-line identifier list is indented after the
// first linebreak encountered.
func (p *printer) identList(list []*ast.Ident, indent bool) {
	// convert into an expression list so we can re-use exprList formatting
	xlist := make([]ast.Expr, len(list))
	for i, x := range list {
		xlist[i] = x
	}
	var mode exprListMode
	if !indent {
		mode = noIndent
	}
	p.exprList(token.NoPos, xlist, 1, mode, token.NoPos)
}

// Print a list of expressions. If the list spans multiple
// source lines, the original line breaks are respected between
// expressions.
//
// TODO(gri) Consider rewriting this to be independent of []ast.Expr
//           so that we can use the algorithm for any kind of list
//           (e.g., pass list via a channel over which to range).
func (p *printer) exprList(prev0 token.Pos, list []ast.Expr, depth int, mode exprListMode, next0 token.Pos) {
	if len(list) == 0 {
		return
	}

	prev := p.posFor(prev0)
	next := p.posFor(next0)
	line := p.lineFor(list[0].Pos())
	endLine := p.lineFor(list[len(list)-1].End())

	if prev.IsValid() && prev.Line == line && line == endLine {
		// all list entries on a single line
		for i, x := range list {
			if i > 0 {
				// use position of expression following the comma as
				// comma position for correct comment placement
				p.print(x.Pos(), token.COMMA, blank)
			}
			p.expr0(x, depth)
		}
		return
	}

	// list entries span multiple lines;
	// use source code positions to guide line breaks

	// don't add extra indentation if noIndent is set;
	// i.e., pretend that the first line is already indented
	ws := ignore
	if mode&noIndent == 0 {
		ws = indent
	}

	// the first linebreak is always a formfeed since this section must not
	// depend on any previous formatting
	prevBreak := -1 // index of last expression that was followed by a linebreak
	if prev.IsValid() && prev.Line < line && p.linebreak(line, 0, ws, true) {
		ws = ignore
		prevBreak = 0
	}

	// initialize expression/key size: a zero value indicates expr/key doesn't fit on a single line
	size := 0

	// print all list elements
	prevLine := prev.Line
	for i, x := range list {
		line = p.lineFor(x.Pos())

		// determine if the next linebreak, if any, needs to use formfeed:
		// in general, use the entire node size to make the decision; for
		// key:value expressions, use the key size
		// TODO(gri) for a better result, should probably incorporate both
		//           the key and the node size into the decision process
		useFF := true

		// determine element size: all bets are off if we don't have
		// position information for the previous and next token (likely
		// generated code - simply ignore the size in this case by setting
		// it to 0)
		prevSize := size
		const infinity = 1e6 // larger than any source line
		size = p.nodeSize(x, infinity)
		pair, isPair := x.(*ast.KeyValueExpr)
		if size <= infinity && prev.IsValid() && next.IsValid() {
			// x fits on a single line
			if isPair {
				size = p.nodeSize(pair.Key, infinity) // size <= infinity
			}
		} else {
			// size too large or we don't have good layout information
			size = 0
		}

		// if the previous line and the current line had single-
		// line-expressions and the key sizes are small or the
		// the ratio between the key sizes does not exceed a
		// threshold, align columns and do not use formfeed
		if prevSize > 0 && size > 0 {
			const smallSize = 20
			if prevSize <= smallSize && size <= smallSize {
				useFF = false
			} else {
				const r = 4 // threshold
				ratio := float64(size) / float64(prevSize)
				useFF = ratio <= 1.0/r || r <= ratio
			}
		}

		needsLinebreak := 0 < prevLine && prevLine < line
		if i > 0 {
			// use position of expression following the comma as
			// comma position for correct comment placement, but
			// only if the expression is on the same line
			if !needsLinebreak {
				p.print(x.Pos())
			}
			p.print(token.COMMA)
			needsBlank := true
			if needsLinebreak {
				// lines are broken using newlines so comments remain aligned
				// unless forceFF is set or there are multiple expressions on
				// the same line in which case formfeed is used
				if p.linebreak(line, 0, ws, useFF || prevBreak+1 < i) {
					ws = ignore
					prevBreak = i
					needsBlank = false // we got a line break instead
				}
			}
			if needsBlank {
				p.print(blank)
			}
		}

		if len(list) > 1 && isPair && size > 0 && needsLinebreak {
			// we have a key:value expression that fits onto one line
			// and it's not on the same line as the prior expression:
			// use a column for the key such that consecutive entries
			// can align if possible
			// (needsLinebreak is set if we started a new line before)
			p.expr(pair.Key)
			p.print(pair.Colon, token.COLON, vtab)
			p.expr(pair.Value)
		} else {
			p.expr0(x, depth)
		}

		prevLine = line
	}

	if mode&commaTerm != 0 && next.IsValid() && p.pos.Line < next.Line {
		// print a terminating comma if the next token is on a new line
		p.print(token.COMMA)
		if ws == ignore && mode&noIndent == 0 {
			// unindent if we indented
			p.print(unindent)
		}
		p.print(formfeed) // terminating comma needs a line break to look good
		return
	}

	if ws == ignore && mode&noIndent == 0 {
		// unindent if we indented
		p.print(unindent)
	}
}

func (p *printer) parameters(fields *ast.FieldList) {
	p.print(fields.Opening, token.LPAREN)
	if len(fields.List) > 0 {
		prevLine := p.lineFor(fields.Opening)
		ws := indent
		for i, par := range fields.List {
			// determine par begin and end line (may be different
			// if there are multiple parameter names for this par
			// or the type is on a separate line)
			var parLineBeg int
			if len(par.Names) > 0 {
				parLineBeg = p.lineFor(par.Names[0].Pos())
			} else {
				parLineBeg = p.lineFor(par.Type.Pos())
			}
			var parLineEnd = p.lineFor(par.Type.End())
			// separating "," if needed
			needsLinebreak := 0 < prevLine && prevLine < parLineBeg
			if i > 0 {
				// use position of parameter following the comma as
				// comma position for correct comma placement, but
				// only if the next parameter is on the same line
				if !needsLinebreak {
					p.print(par.Pos())
				}
				p.print(token.COMMA)
			}
			// separator if needed (linebreak or blank)
			if needsLinebreak && p.linebreak(parLineBeg, 0, ws, true) {
				// break line if the opening "(" or previous parameter ended on a different line
				ws = ignore
			} else if i > 0 {
				p.print(blank)
			}
			// parameter names
			if len(par.Names) > 0 {
				// Very subtle: If we indented before (ws == ignore), identList
				// won't indent again. If we didn't (ws == indent), identList will
				// indent if the identList spans multiple lines, and it will outdent
				// again at the end (and still ws == indent). Thus, a subsequent indent
				// by a linebreak call after a type, or in the next multi-line identList
				// will do the right thing.
				p.identList(par.Names, ws == indent)
				p.print(blank)
			}
			// parameter type
			p.expr(stripParensAlways(par.Type))
			prevLine = parLineEnd
		}
		// if the closing ")" is on a separate line from the last parameter,
		// print an additional "," and line break
		if closing := p.lineFor(fields.Closing); 0 < prevLine && prevLine < closing {
			p.print(token.COMMA)
			p.linebreak(closing, 0, ignore, true)
		}
		// unindent if we indented
		if ws == ignore {
			p.print(unindent)
		}
	}
	p.print(fields.Closing, token.RPAREN)
}

func (p *printer) signature(params, result *ast.FieldList) {
	if params != nil {
		p.parameters(params)
	} else {
		p.print(token.LPAREN, token.RPAREN)
	}
	n := result.NumFields()
	if n > 0 {
		// result != nil
		p.print(blank)
		if n == 1 && result.List[0].Names == nil {
			// single anonymous result; no ()'s
			p.expr(stripParensAlways(result.List[0].Type))
			return
		}
		p.parameters(result)
	}
}

func identListSize(list []*ast.Ident, maxSize int) (size int) {
	for i, x := range list {
		if i > 0 {
			size += len(", ")
		}
		size += utf8.RuneCountInString(x.Name)
		if size >= maxSize {
			break
		}
	}
	return
}

func (p *printer) isOneLineFieldList(list []*ast.Field) bool {
	if len(list) != 1 {
		return false // allow only one field
	}
	f := list[0]
	if f.Tag != nil || f.Comment != nil {
		return false // don't allow tags or comments
	}
	// only name(s) and type
	const maxSize = 30 // adjust as appropriate, this is an approximate value
	namesSize := identListSize(f.Names, maxSize)
	if namesSize > 0 {
		namesSize = 1 // blank between names and types
	}
	typeSize := p.nodeSize(f.Type, maxSize)
	return namesSize+typeSize <= maxSize
}

func (p *printer) setLineComment(text string) {
	p.setComment(&ast.CommentGroup{List: []*ast.Comment{{Slash: token.NoPos, Text: text}}})
}

func (p *printer) fieldList(fields *ast.FieldList, isStruct, isIncomplete bool) {
	lbrace := fields.Opening
	list := fields.List
	rbrace := fields.Closing
	hasComments := isIncomplete || p.commentBefore(p.posFor(rbrace))
	srcIsOneLine := lbrace.IsValid() && rbrace.IsValid() && p.lineFor(lbrace) == p.lineFor(rbrace)

	if !hasComments && srcIsOneLine {
		// possibly a one-line struct/interface
		if len(list) == 0 {
			// no blank between keyword and {} in this case
			p.print(lbrace, token.LBRACE, rbrace, token.RBRACE)
			return
		} else if isStruct && p.isOneLineFieldList(list) { // for now ignore interfaces
			// small enough - print on one line
			// (don't use identList and ignore source line breaks)
			p.print(lbrace, token.LBRACE, blank)
			f := list[0]
			for i, x := range f.Names {
				if i > 0 {
					// no comments so no need for comma position
					p.print(token.COMMA, blank)
				}
				p.expr(x)
			}
			if len(f.Names) > 0 {
				p.print(blank)
			}
			p.expr(f.Type)
			p.print(blank, rbrace, token.RBRACE)
			return
		}
	}
	// hasComments || !srcIsOneLine

	p.print(blank, lbrace, token.LBRACE, indent)
	if hasComments || len(list) > 0 {
		p.print(formfeed)
	}

	if isStruct {

		sep := vtab
		if len(list) == 1 {
			sep = blank
		}
		var line int
		for i, f := range list {
			if i > 0 {
				p.linebreak(p.lineFor(f.Pos()), 1, ignore, p.linesFrom(line) > 0)
			}
			extraTabs := 0
			p.setComment(f.Doc)
			p.recordLine(&line)
			if len(f.Names) > 0 {
				// named fields
				p.identList(f.Names, false)
				p.print(sep)
				p.expr(f.Type)
				extraTabs = 1
			} else {
				// anonymous field
				p.expr(f.Type)
				extraTabs = 2
			}
			if f.Tag != nil {
				if len(f.Names) > 0 && sep == vtab {
					p.print(sep)
				}
				p.print(sep)
				p.expr(f.Tag)
				extraTabs = 0
			}
			if f.Comment != nil {
				for ; extraTabs > 0; extraTabs-- {
					p.print(sep)
				}
				p.setComment(f.Comment)
			}
		}
		if isIncomplete {
			if len(list) > 0 {
				p.print(formfeed)
			}
			p.flush(p.posFor(rbrace), token.RBRACE) // make sure we don't lose the last line comment
			p.setLineComment("// contains filtered or unexported fields")
		}

	} else { // interface

		var line int
		for i, f := range list {
			if i > 0 {
				p.linebreak(p.lineFor(f.Pos()), 1, ignore, p.linesFrom(line) > 0)
			}
			p.setComment(f.Doc)
			p.recordLine(&line)
			if ftyp, isFtyp := f.Type.(*ast.FuncType); isFtyp {
				// method
				p.expr(f.Names[0])
				p.signature(ftyp.Params, ftyp.Results)
			} else {
				// embedded interface
				p.expr(f.Type)
			}
			p.setComment(f.Comment)
		}
		if isIncomplete {
			if len(list) > 0 {
				p.print(formfeed)
			}
			p.flush(p.posFor(rbrace), token.RBRACE) // make sure we don't lose the last line comment
			p.setLineComment("// contains filtered or unexported methods")
		}

	}
	p.print(unindent, formfeed, rbrace, token.RBRACE)
}

// ----------------------------------------------------------------------------
// Expressions

func walkBinary(e *ast.BinaryExpr) (has4, has5 bool, maxProblem int) {
	switch e.Op.Precedence() {
	case 4:
		has4 = true
	case 5:
		has5 = true
	}

	switch l := e.X.(type) {
	case *ast.BinaryExpr:
		if l.Op.Precedence() < e.Op.Precedence() {
			// parens will be inserted.
			// pretend this is an *ast.ParenExpr and do nothing.
			break
		}
		h4, h5, mp := walkBinary(l)
		has4 = has4 || h4
		has5 = has5 || h5
		if maxProblem < mp {
			maxProblem = mp
		}
	}

	switch r := e.Y.(type) {
	case *ast.BinaryExpr:
		if r.Op.Precedence() <= e.Op.Precedence() {
			// parens will be inserted.
			// pretend this is an *ast.ParenExpr and do nothing.
			break
		}
		h4, h5, mp := walkBinary(r)
		has4 = has4 || h4
		has5 = has5 || h5
		if maxProblem < mp {
			maxProblem = mp
		}

	case *ast.StarExpr:
		if e.Op == token.QUO { // `*/`
			maxProblem = 5
		}

	case *ast.UnaryExpr:
		switch e.Op.String() + r.Op.String() {
		case "/*", "&&", "&^":
			maxProblem = 5
		case "++", "--":
			if maxProblem < 4 {
				maxProblem = 4
			}
		}
	}
	return
}

func cutoff(e *ast.BinaryExpr, depth int) int {
	has4, has5, maxProblem := walkBinary(e)
	if maxProblem > 0 {
		return maxProblem + 1
	}
	if has4 && has5 {
		if depth == 1 {
			return 5
		}
		return 4
	}
	if depth == 1 {
		return 6
	}
	return 4
}

func diffPrec(expr ast.Expr, prec int) int {
	x, ok := expr.(*ast.BinaryExpr)
	if !ok || prec != x.Op.Precedence() {
		return 1
	}
	return 0
}

func reduceDepth(depth int) int {
	depth--
	if depth < 1 {
		depth = 1
	}
	return depth
}

// Format the binary expression: decide the cutoff and then format.
// Let's call depth == 1 Normal mode, and depth > 1 Compact mode.
// (Algorithm suggestion by Russ Cox.)
//
// The precedences are:
//	5             *  /  %  <<  >>  &  &^
//	4             +  -  |  ^
//	3             ==  !=  <  <=  >  >=
//	2             &&
//	1             ||
//
// The only decision is whether there will be spaces around levels 4 and 5.
// There are never spaces at level 6 (unary), and always spaces at levels 3 and below.
//
// To choose the cutoff, look at the whole expression but excluding primary
// expressions (function calls, parenthesized exprs), and apply these rules:
//
//	1) If there is a binary operator with a right side unary operand
//	   that would clash without a space, the cutoff must be (in order):
//
//		/*	6
//		&&	6
//		&^	6
//		++	5
//		--	5
//
//         (Comparison operators always have spaces around them.)
//
//	2) If there is a mix of level 5 and level 4 operators, then the cutoff
//	   is 5 (use spaces to distinguish precedence) in Normal mode
//	   and 4 (never use spaces) in Compact mode.
//
//	3) If there are no level 4 operators or no level 5 operators, then the
//	   cutoff is 6 (always use spaces) in Normal mode
//	   and 4 (never use spaces) in Compact mode.
//
func (p *printer) binaryExpr(x *ast.BinaryExpr, prec1, cutoff, depth int) {
	prec := x.Op.Precedence()
	if prec < prec1 {
		// parenthesis needed
		// Note: The parser inserts an ast.ParenExpr node; thus this case
		//       can only occur if the AST is created in a different way.
		p.print(token.LPAREN)
		p.expr0(x, reduceDepth(depth)) // parentheses undo one level of depth
		p.print(token.RPAREN)
		return
	}

	printBlank := prec < cutoff

	ws := indent
	p.expr1(x.X, prec, depth+diffPrec(x.X, prec))
	if printBlank {
		p.print(blank)
	}
	xline := p.pos.Line // before the operator (it may be on the next line!)
	yline := p.lineFor(x.Y.Pos())
	p.print(x.OpPos, x.Op)
	if xline != yline && xline > 0 && yline > 0 {
		// at least one line break, but respect an extra empty line
		// in the source
		if p.linebreak(yline, 1, ws, true) {
			ws = ignore
			printBlank = false // no blank after line break
		}
	}
	if printBlank {
		p.print(blank)
	}
	p.expr1(x.Y, prec+1, depth+1)
	if ws == ignore {
		p.print(unindent)
	}
}

func isBinary(expr ast.Expr) bool {
	_, ok := expr.(*ast.BinaryExpr)
	return ok
}

func (p *printer) expr1(expr ast.Expr, prec1, depth int) {
	p.print(expr.Pos())

	switch x := expr.(type) {
	case *ast.BadExpr:
		p.print("BadExpr")

	case *ast.Ident:
		p.print(x)

	case *ast.BinaryExpr:
		if depth < 1 {
			p.internalError("depth < 1:", depth)
			depth = 1
		}
		p.binaryExpr(x, prec1, cutoff(x, depth), depth)

	case *ast.KeyValueExpr:
		p.expr(x.Key)
		p.print(x.Colon, token.COLON, blank)
		p.expr(x.Value)

	case *ast.StarExpr:
		const prec = token.UnaryPrec
		if prec < prec1 {
			// parenthesis needed
			p.print(token.LPAREN)
			p.print(token.MUL)
			p.expr(x.X)
			p.print(token.RPAREN)
		} else {
			// no parenthesis needed
			p.print(token.MUL)
			p.expr(x.X)
		}

	case *ast.UnaryExpr:
		const prec = token.UnaryPrec
		if prec < prec1 {
			// parenthesis needed
			p.print(token.LPAREN)
			p.expr(x)
			p.print(token.RPAREN)
		} else {
			// no parenthesis needed
			p.print(x.Op)
			if x.Op == token.RANGE {
				// TODO(gri) Remove this code if it cannot be reached.
				p.print(blank)
			}
			p.expr1(x.X, prec, depth)
		}

	case *ast.BasicLit:
		p.print(x)

	case *ast.FuncLit:
		p.expr(x.Type)
		p.adjBlock(p.distanceFrom(x.Type.Pos()), blank, x.Body)

	case *ast.ParenExpr:
		if _, hasParens := x.X.(*ast.ParenExpr); hasParens {
			// don't print parentheses around an already parenthesized expression
			// TODO(gri) consider making this more general and incorporate precedence levels
			p.expr0(x.X, depth)
		} else {
			p.print(token.LPAREN)
			p.expr0(x.X, reduceDepth(depth)) // parentheses undo one level of depth
			p.print(x.Rparen, token.RPAREN)
		}

	case *ast.SelectorExpr:
		p.selectorExpr(x, depth, false)

	case *ast.TypeAssertExpr:
		p.expr1(x.X, token.HighestPrec, depth)
		p.print(token.PERIOD, x.Lparen, token.LPAREN)
		if x.Type != nil {
			p.expr(x.Type)
		} else {
			p.print(token.TYPE)
		}
		p.print(x.Rparen, token.RPAREN)

	case *ast.IndexExpr:
		// TODO(gri): should treat[] like parentheses and undo one level of depth
		p.expr1(x.X, token.HighestPrec, 1)
		p.print(x.Lbrack, token.LBRACK)
		p.expr0(x.Index, depth+1)
		p.print(x.Rbrack, token.RBRACK)

	case *ast.SliceExpr:
		// TODO(gri): should treat[] like parentheses and undo one level of depth
		p.expr1(x.X, token.HighestPrec, 1)
		p.print(x.Lbrack, token.LBRACK)
		indices := []ast.Expr{x.Low, x.High}
		if x.Max != nil {
			indices = append(indices, x.Max)
		}
		for i, y := range indices {
			if i > 0 {
				// blanks around ":" if both sides exist and either side is a binary expression
				// TODO(gri) once we have committed a variant of a[i:j:k] we may want to fine-
				//           tune the formatting here
				x := indices[i-1]
				if depth <= 1 && x != nil && y != nil && (isBinary(x) || isBinary(y)) {
					p.print(blank, token.COLON, blank)
				} else {
					p.print(token.COLON)
				}
			}
			if y != nil {
				p.expr0(y, depth+1)
			}
		}
		p.print(x.Rbrack, token.RBRACK)

	case *ast.CallExpr:
		if len(x.Args) > 1 {
			depth++
		}
		var wasIndented bool
		if _, ok := x.Fun.(*ast.FuncType); ok {
			// conversions to literal function types require parentheses around the type
			p.print(token.LPAREN)
			wasIndented = p.possibleSelectorExpr(x.Fun, token.HighestPrec, depth)
			p.print(token.RPAREN)
		} else {
			wasIndented = p.possibleSelectorExpr(x.Fun, token.HighestPrec, depth)
		}
		p.print(x.Lparen, token.LPAREN)
		if x.Ellipsis.IsValid() {
			p.exprList(x.Lparen, x.Args, depth, 0, x.Ellipsis)
			p.print(x.Ellipsis, token.ELLIPSIS)
			if x.Rparen.IsValid() && p.lineFor(x.Ellipsis) < p.lineFor(x.Rparen) {
				p.print(token.COMMA, formfeed)
			}
		} else {
			p.exprList(x.Lparen, x.Args, depth, commaTerm, x.Rparen)
		}
		p.print(x.Rparen, token.RPAREN)
		if wasIndented {
			p.print(unindent)
		}

	case *ast.CompositeLit:
		// composite literal elements that are composite literals themselves may have the type omitted
		if x.Type != nil {
			p.expr1(x.Type, token.HighestPrec, depth)
		}
		p.print(x.Lbrace, token.LBRACE)
		p.exprList(x.Lbrace, x.Elts, 1, commaTerm, x.Rbrace)
		// do not insert extra line break following a /*-style comment
		// before the closing '}' as it might break the code if there
		// is no trailing ','
		mode := noExtraLinebreak
		// do not insert extra blank following a /*-style comment
		// before the closing '}' unless the literal is empty
		if len(x.Elts) > 0 {
			mode |= noExtraBlank
		}
		p.print(mode, x.Rbrace, token.RBRACE, mode)

	case *ast.Ellipsis:
		p.print(token.ELLIPSIS)
		if x.Elt != nil {
			p.expr(x.Elt)
		}

	case *ast.ArrayType:
		p.print(token.LBRACK)
		if x.Len != nil {
			p.expr(x.Len)
		}
		p.print(token.RBRACK)
		p.expr(x.Elt)

	case *ast.StructType:
		p.print(token.STRUCT)
		p.fieldList(x.Fields, true, x.Incomplete)

	case *ast.FuncType:
		p.print(token.FUNC)
		p.signature(x.Params, x.Results)

	case *ast.InterfaceType:
		p.print(token.INTERFACE)
		p.fieldList(x.Methods, false, x.Incomplete)

	case *ast.MapType:
		p.print(token.MAP, token.LBRACK)
		p.expr(x.Key)
		p.print(token.RBRACK)
		p.expr(x.Value)

	case *ast.ChanType:
		switch x.Dir {
		case ast.SEND | ast.RECV:
			p.print(token.CHAN)
		case ast.RECV:
			p.print(token.ARROW, token.CHAN) // x.Arrow and x.Pos() are the same
		case ast.SEND:
			p.print(token.CHAN, x.Arrow, token.ARROW)
		}
		p.print(blank)
		p.expr(x.Value)

	default:
		panic("unreachable")
	}

	return
}

func (p *printer) possibleSelectorExpr(expr ast.Expr, prec1, depth int) bool {
	if x, ok := expr.(*ast.SelectorExpr); ok {
		return p.selectorExpr(x, depth, true)
	}
	p.expr1(expr, prec1, depth)
	return false
}

// selectorExpr handles an *ast.SelectorExpr node and returns whether x spans
// multiple lines.
func (p *printer) selectorExpr(x *ast.SelectorExpr, depth int, isMethod bool) bool {
	p.expr1(x.X, token.HighestPrec, depth)
	p.print(token.PERIOD)
	if line := p.lineFor(x.Sel.Pos()); p.pos.IsValid() && p.pos.Line < line {
		p.print(indent, newline, x.Sel.Pos(), x.Sel)
		if !isMethod {
			p.print(unindent)
		}
		return true
	}
	p.print(x.Sel.Pos(), x.Sel)
	return false
}

func (p *printer) expr0(x ast.Expr, depth int) {
	p.expr1(x, token.LowestPrec, depth)
}

func (p *printer) expr(x ast.Expr) {
	const depth = 1
	p.expr1(x, token.LowestPrec, depth)
}

// ----------------------------------------------------------------------------
// Statements

// Print the statement list indented, but without a newline after the last statement.
// Extra line breaks between statements in the source are respected but at most one
// empty line is printed between statements.
func (p *printer) stmtList(list []ast.Stmt, nindent int, nextIsRBrace bool) {
	if nindent > 0 {
		p.print(indent)
	}
	var line int
	i := 0
	for _, s := range list {
		// ignore empty statements (was issue 3466)
		if _, isEmpty := s.(*ast.EmptyStmt); !isEmpty {
			// nindent == 0 only for lists of switch/select case clauses;
			// in those cases each clause is a new section
			if len(p.output) > 0 {
				// only print line break if we are not at the beginning of the output
				// (i.e., we are not printing only a partial program)
				p.linebreak(p.lineFor(s.Pos()), 1, ignore, i == 0 || nindent == 0 || p.linesFrom(line) > 0)
			}
			p.recordLine(&line)
			p.stmt(s, nextIsRBrace && i == len(list)-1)
			// labeled statements put labels on a separate line, but here
			// we only care about the start line of the actual statement
			// without label - correct line for each label
			for t := s; ; {
				lt, _ := t.(*ast.LabeledStmt)
				if lt == nil {
					break
				}
				line++
				t = lt.Stmt
			}
			i++
		}
	}
	if nindent > 0 {
		p.print(unindent)
	}
}

// block prints an *ast.BlockStmt; it always spans at least two lines.
func (p *printer) block(b *ast.BlockStmt, nindent int) {
	p.print(b.Lbrace, token.LBRACE)
	p.stmtList(b.List, nindent, true)
	p.linebreak(p.lineFor(b.Rbrace), 1, ignore, true)
	p.print(b.Rbrace, token.RBRACE)
}

func isTypeName(x ast.Expr) bool {
	switch t := x.(type) {
	case *ast.Ident:
		return true
	case *ast.SelectorExpr:
		return isTypeName(t.X)
	}
	return false
}

func stripParens(x ast.Expr) ast.Expr {
	if px, strip := x.(*ast.ParenExpr); strip {
		// parentheses must not be stripped if there are any
		// unparenthesized composite literals starting with
		// a type name
		ast.Inspect(px.X, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.ParenExpr:
				// parentheses protect enclosed composite literals
				return false
			case *ast.CompositeLit:
				if isTypeName(x.Type) {
					strip = false // do not strip parentheses
				}
				return false
			}
			// in all other cases, keep inspecting
			return true
		})
		if strip {
			return stripParens(px.X)
		}
	}
	return x
}

func stripParensAlways(x ast.Expr) ast.Expr {
	if x, ok := x.(*ast.ParenExpr); ok {
		return stripParensAlways(x.X)
	}
	return x
}

func (p *printer) controlClause(isForStmt bool, init ast.Stmt, expr ast.Expr, post ast.Stmt) {
	p.print(blank)
	needsBlank := false
	if init == nil && post == nil {
		// no semicolons required
		if expr != nil {
			p.expr(stripParens(expr))
			needsBlank = true
		}
	} else {
		// all semicolons required
		// (they are not separators, print them explicitly)
		if init != nil {
			p.stmt(init, false)
		}
		p.print(token.SEMICOLON, blank)
		if expr != nil {
			p.expr(stripParens(expr))
			needsBlank = true
		}
		if isForStmt {
			p.print(token.SEMICOLON, blank)
			needsBlank = false
			if post != nil {
				p.stmt(post, false)
				needsBlank = true
			}
		}
	}
	if needsBlank {
		p.print(blank)
	}
}

// indentList reports whether an expression list would look better if it
// were indented wholesale (starting with the very first element, rather
// than starting at the first line break).
//
func (p *printer) indentList(list []ast.Expr) bool {
	// Heuristic: indentList returns true if there are more than one multi-
	// line element in the list, or if there is any element that is not
	// starting on the same line as the previous one ends.
	if len(list) >= 2 {
		var b = p.lineFor(list[0].Pos())
		var e = p.lineFor(list[len(list)-1].End())
		if 0 < b && b < e {
			// list spans multiple lines
			n := 0 // multi-line element count
			line := b
			for _, x := range list {
				xb := p.lineFor(x.Pos())
				xe := p.lineFor(x.End())
				if line < xb {
					// x is not starting on the same
					// line as the previous one ended
					return true
				}
				if xb < xe {
					// x is a multi-line element
					n++
				}
				line = xe
			}
			return n > 1
		}
	}
	return false
}

func (p *printer) stmt(stmt ast.Stmt, nextIsRBrace bool) {
	p.print(stmt.Pos())

	switch s := stmt.(type) {
	case *ast.BadStmt:
		p.print("BadStmt")

	case *ast.DeclStmt:
		p.decl(s.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		// a "correcting" unindent immediately following a line break
		// is applied before the line break if there is no comment
		// between (see writeWhitespace)
		p.print(unindent)
		p.expr(s.Label)
		p.print(s.Colon, token.COLON, indent)
		if e, isEmpty := s.Stmt.(*ast.EmptyStmt); isEmpty {
			if !nextIsRBrace {
				p.print(newline, e.Pos(), token.SEMICOLON)
				break
			}
		} else {
			p.linebreak(p.lineFor(s.Stmt.Pos()), 1, ignore, true)
		}
		p.stmt(s.Stmt, nextIsRBrace)

	case *ast.ExprStmt:
		const depth = 1
		p.expr0(s.X, depth)

	case *ast.SendStmt:
		const depth = 1
		p.expr0(s.Chan, depth)
		p.print(blank, s.Arrow, token.ARROW, blank)
		p.expr0(s.Value, depth)

	case *ast.IncDecStmt:
		const depth = 1
		p.expr0(s.X, depth+1)
		p.print(s.TokPos, s.Tok)

	case *ast.AssignStmt:
		var depth = 1
		if len(s.Lhs) > 1 && len(s.Rhs) > 1 {
			depth++
		}
		p.exprList(s.Pos(), s.Lhs, depth, 0, s.TokPos)
		p.print(blank, s.TokPos, s.Tok, blank)
		p.exprList(s.TokPos, s.Rhs, depth, 0, token.NoPos)

	case *ast.GoStmt:
		p.print(token.GO, blank)
		p.expr(s.Call)

	case *ast.DeferStmt:
		p.print(token.DEFER, blank)
		p.expr(s.Call)

	case *ast.ReturnStmt:
		p.print(token.RETURN)
		if s.Results != nil {
			p.print(blank)
			// Use indentList heuristic to make corner cases look
			// better (issue 1207). A more systematic approach would
			// always indent, but this would cause significant
			// reformatting of the code base and not necessarily
			// lead to more nicely formatted code in general.
			if p.indentList(s.Results) {
				p.print(indent)
				p.exprList(s.Pos(), s.Results, 1, noIndent, token.NoPos)
				p.print(unindent)
			} else {
				p.exprList(s.Pos(), s.Results, 1, 0, token.NoPos)
			}
		}

	case *ast.BranchStmt:
		p.print(s.Tok)
		if s.Label != nil {
			p.print(blank)
			p.expr(s.Label)
		}

	case *ast.BlockStmt:
		p.block(s, 1)

	case *ast.IfStmt:
		p.print(token.IF)
		p.controlClause(false, s.Init, s.Cond, nil)
		p.block(s.Body, 1)
		if s.Else != nil {
			p.print(blank, token.ELSE, blank)
			switch s.Else.(type) {
			case *ast.BlockStmt, *ast.IfStmt:
				p.stmt(s.Else, nextIsRBrace)
			default:
				// This can only happen with an incorrectly
				// constructed AST. Permit it but print so
				// that it can be parsed without errors.
				p.print(token.LBRACE, indent, formfeed)
				p.stmt(s.Else, true)
				p.print(unindent, formfeed, token.RBRACE)
			}
		}

	case *ast.CaseClause:
		if s.List != nil {
			p.print(token.CASE, blank)
			p.exprList(s.Pos(), s.List, 1, 0, s.Colon)
		} else {
			p.print(token.DEFAULT)
		}
		p.print(s.Colon, token.COLON)
		p.stmtList(s.Body, 1, nextIsRBrace)

	case *ast.SwitchStmt:
		p.print(token.SWITCH)
		p.controlClause(false, s.Init, s.Tag, nil)
		p.block(s.Body, 0)

	case *ast.TypeSwitchStmt:
		p.print(token.SWITCH)
		if s.Init != nil {
			p.print(blank)
			p.stmt(s.Init, false)
			p.print(token.SEMICOLON)
		}
		p.print(blank)
		p.stmt(s.Assign, false)
		p.print(blank)
		p.block(s.Body, 0)

	case *ast.CommClause:
		if s.Comm != nil {
			p.print(token.CASE, blank)
			p.stmt(s.Comm, false)
		} else {
			p.print(token.DEFAULT)
		}
		p.print(s.Colon, token.COLON)
		p.stmtList(s.Body, 1, nextIsRBrace)

	case *ast.SelectStmt:
		p.print(token.SELECT, blank)
		body := s.Body
		if len(body.List) == 0 && !p.commentBefore(p.posFor(body.Rbrace)) {
			// print empty select statement w/o comments on one line
			p.print(body.Lbrace, token.LBRACE, body.Rbrace, token.RBRACE)
		} else {
			p.block(body, 0)
		}

	case *ast.ForStmt:
		p.print(token.FOR)
		p.controlClause(true, s.Init, s.Cond, s.Post)
		p.block(s.Body, 1)

	case *ast.RangeStmt:
		p.print(token.FOR, blank)
		if s.Key != nil {
			p.expr(s.Key)
			if s.Value != nil {
				// use position of value following the comma as
				// comma position for correct comment placement
				p.print(s.Value.Pos(), token.COMMA, blank)
				p.expr(s.Value)
			}
			p.print(blank, s.TokPos, s.Tok, blank)
		}
		p.print(token.RANGE, blank)
		p.expr(stripParens(s.X))
		p.print(blank)
		p.block(s.Body, 1)

	default:
		panic("unreachable")
	}

	return
}

// ----------------------------------------------------------------------------
// Declarations

// The keepTypeColumn function determines if the type column of a series of
// consecutive const or var declarations must be kept, or if initialization
// values (V) can be placed in the type column (T) instead. The i'th entry
// in the result slice is true if the type column in spec[i] must be kept.
//
// For example, the declaration:
//
//	const (
//		foobar int = 42 // comment
//		x          = 7  // comment
//		foo
//              bar = 991
//	)
//
// leads to the type/values matrix below. A run of value columns (V) can
// be moved into the type column if there is no type for any of the values
// in that column (we only move entire columns so that they align properly).
//
//	matrix        formatted     result
//                    matrix
//	T  V    ->    T  V     ->   true      there is a T and so the type
//	-  V          -  V          true      column must be kept
//	-  -          -  -          false
//	-  V          V  -          false     V is moved into T column
//
func keepTypeColumn(specs []ast.Spec) []bool {
	m := make([]bool, len(specs))

	populate := func(i, j int, keepType bool) {
		if keepType {
			for ; i < j; i++ {
				m[i] = true
			}
		}
	}

	i0 := -1 // if i0 >= 0 we are in a run and i0 is the start of the run
	var keepType bool
	for i, s := range specs {
		t := s.(*ast.ValueSpec)
		if t.Values != nil {
			if i0 < 0 {
				// start of a run of ValueSpecs with non-nil Values
				i0 = i
				keepType = false
			}
		} else {
			if i0 >= 0 {
				// end of a run
				populate(i0, i, keepType)
				i0 = -1
			}
		}
		if t.Type != nil {
			keepType = true
		}
	}
	if i0 >= 0 {
		// end of a run
		populate(i0, len(specs), keepType)
	}

	return m
}

func (p *printer) valueSpec(s *ast.ValueSpec, keepType bool) {
	p.setComment(s.Doc)
	p.identList(s.Names, false) // always present
	extraTabs := 3
	if s.Type != nil || keepType {
		p.print(vtab)
		extraTabs--
	}
	if s.Type != nil {
		p.expr(s.Type)
	}
	if s.Values != nil {
		p.print(vtab, token.ASSIGN, blank)
		p.exprList(token.NoPos, s.Values, 1, 0, token.NoPos)
		extraTabs--
	}
	if s.Comment != nil {
		for ; extraTabs > 0; extraTabs-- {
			p.print(vtab)
		}
		p.setComment(s.Comment)
	}
}

func sanitizeImportPath(lit *ast.BasicLit) *ast.BasicLit {
	// Note: An unmodified AST generated by go/parser will already
	// contain a backward- or double-quoted path string that does
	// not contain any invalid characters, and most of the work
	// here is not needed. However, a modified or generated AST
	// may possibly contain non-canonical paths. Do the work in
	// all cases since it's not too hard and not speed-critical.

	// if we don't have a proper string, be conservative and return whatever we have
	if lit.Kind != token.STRING {
		return lit
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return lit
	}

	// if the string is an invalid path, return whatever we have
	//
	// spec: "Implementation restriction: A compiler may restrict
	// ImportPaths to non-empty strings using only characters belonging
	// to Unicode's L, M, N, P, and S general categories (the Graphic
	// characters without spaces) and may also exclude the characters
	// !"#$%&'()*,:;<=>?[\]^`{|} and the Unicode replacement character
	// U+FFFD."
	if s == "" {
		return lit
	}
	const illegalChars = `!"#$%&'()*,:;<=>?[\]^{|}` + "`\uFFFD"
	for _, r := range s {
		if !unicode.IsGraphic(r) || unicode.IsSpace(r) || strings.ContainsRune(illegalChars, r) {
			return lit
		}
	}

	// otherwise, return the double-quoted path
	s = strconv.Quote(s)
	if s == lit.Value {
		return lit // nothing wrong with lit
	}
	return &ast.BasicLit{ValuePos: lit.ValuePos, Kind: token.STRING, Value: s}
}

// The parameter n is the number of specs in the group. If doIndent is set,
// multi-line identifier lists in the spec are indented when the first
// linebreak is encountered.
//
func (p *printer) spec(spec ast.Spec, n int, doIndent bool) {
	switch s := spec.(type) {
	case *ast.ImportSpec:
		p.setComment(s.Doc)
		if s.Name != nil {
			p.expr(s.Name)
			p.print(blank)
		}
		p.expr(sanitizeImportPath(s.Path))
		p.setComment(s.Comment)
		p.print(s.EndPos)

	case *ast.ValueSpec:
		if n != 1 {
			p.internalError("expected n = 1; got", n)
		}
		p.setComment(s.Doc)
		p.identList(s.Names, doIndent) // always present
		if s.Type != nil {
			p.print(blank)
			p.expr(s.Type)
		}
		if s.Values != nil {
			p.print(blank, token.ASSIGN, blank)
			p.exprList(token.NoPos, s.Values, 1, 0, token.NoPos)
		}
		p.setComment(s.Comment)

	case *ast.TypeSpec:
		p.setComment(s.Doc)
		p.expr(s.Name)
		if n == 1 {
			p.print(blank)
		} else {
			p.print(vtab)
		}
		p.expr(s.Type)
		p.setComment(s.Comment)

	default:
		panic("unreachable")
	}
}

func (p *printer) genDecl(d *ast.GenDecl) {
	p.setComment(d.Doc)
	p.print(d.Pos(), d.Tok, blank)

	if d.Lparen.IsValid() {
		// group of parenthesized declarations
		p.print(d.Lparen, token.LPAREN)
		if n := len(d.Specs); n > 0 {
			p.print(indent, formfeed)
			if n > 1 && (d.Tok == token.CONST || d.Tok == token.VAR) {
				// two or more grouped const/var declarations:
				// determine if the type column must be kept
				keepType := keepTypeColumn(d.Specs)
				var line int
				for i, s := range d.Specs {
					if i > 0 {
						p.linebreak(p.lineFor(s.Pos()), 1, ignore, p.linesFrom(line) > 0)
					}
					p.recordLine(&line)
					p.valueSpec(s.(*ast.ValueSpec), keepType[i])
				}
			} else {
				var line int
				for i, s := range d.Specs {
					if i > 0 {
						p.linebreak(p.lineFor(s.Pos()), 1, ignore, p.linesFrom(line) > 0)
					}
					p.recordLine(&line)
					p.spec(s, n, false)
				}
			}
			p.print(unindent, formfeed)
		}
		p.print(d.Rparen, token.RPAREN)

	} else {
		// single declaration
		p.spec(d.Specs[0], 1, true)
	}
}

// nodeSize determines the size of n in chars after formatting.
// The result is <= maxSize if the node fits on one line with at
// most maxSize chars and the formatted output doesn't contain
// any control chars. Otherwise, the result is > maxSize.
//
func (p *printer) nodeSize(n ast.Node, maxSize int) (size int) {
	// nodeSize invokes the printer, which may invoke nodeSize
	// recursively. For deep composite literal nests, this can
	// lead to an exponential algorithm. Remember previous
	// results to prune the recursion (was issue 1628).
	if size, found := p.nodeSizes[n]; found {
		return size
	}

	size = maxSize + 1 // assume n doesn't fit
	p.nodeSizes[n] = size

	// nodeSize computation must be independent of particular
	// style so that we always get the same decision; print
	// in RawFormat
	cfg := Config{Mode: RawFormat}
	var buf bytes.Buffer
	if err := cfg.fprint(&buf, p.fset, n, p.nodeSizes); err != nil {
		return
	}
	if buf.Len() <= maxSize {
		for _, ch := range buf.Bytes() {
			if ch < ' ' {
				return
			}
		}
		size = buf.Len() // n fits
		p.nodeSizes[n] = size
	}
	return
}

// bodySize is like nodeSize but it is specialized for *ast.BlockStmt's.
func (p *printer) bodySize(b *ast.BlockStmt, maxSize int) int {
	pos1 := b.Pos()
	pos2 := b.Rbrace
	if pos1.IsValid() && pos2.IsValid() && p.lineFor(pos1) != p.lineFor(pos2) {
		// opening and closing brace are on different lines - don't make it a one-liner
		return maxSize + 1
	}
	if len(b.List) > 5 {
		// too many statements - don't make it a one-liner
		return maxSize + 1
	}
	// otherwise, estimate body size
	bodySize := p.commentSizeBefore(p.posFor(pos2))
	for i, s := range b.List {
		if bodySize > maxSize {
			break // no need to continue
		}
		if i > 0 {
			bodySize += 2 // space for a semicolon and blank
		}
		bodySize += p.nodeSize(s, maxSize)
	}
	return bodySize
}

// adjBlock prints an "adjacent" block (e.g., a for-loop or function body) following
// a header (e.g., a for-loop control clause or function signature) of given headerSize.
// If the header's and block's size are "small enough" and the block is "simple enough",
// the block is printed on the current line, without line breaks, spaced from the header
// by sep. Otherwise the block's opening "{" is printed on the current line, followed by
// lines for the block's statements and its closing "}".
//
func (p *printer) adjBlock(headerSize int, sep whiteSpace, b *ast.BlockStmt) {
	if b == nil {
		return
	}

	const maxSize = 100
	if headerSize+p.bodySize(b, maxSize) <= maxSize {
		p.print(sep, b.Lbrace, token.LBRACE)
		if len(b.List) > 0 {
			p.print(blank)
			for i, s := range b.List {
				if i > 0 {
					p.print(token.SEMICOLON, blank)
				}
				p.stmt(s, i == len(b.List)-1)
			}
			p.print(blank)
		}
		p.print(noExtraLinebreak, b.Rbrace, token.RBRACE, noExtraLinebreak)
		return
	}

	if sep != ignore {
		p.print(blank) // always use blank
	}
	p.block(b, 1)
}

// distanceFrom returns the column difference between from and p.pos (the current
// estimated position) if both are on the same line; if they are on different lines
// (or unknown) the result is infinity.
func (p *printer) distanceFrom(from token.Pos) int {
	if from.IsValid() && p.pos.IsValid() {
		if f := p.posFor(from); f.Line == p.pos.Line {
			return p.pos.Column - f.Column
		}
	}
	return infinity
}

func (p *printer) funcDecl(d *ast.FuncDecl) {
	p.setComment(d.Doc)
	p.print(d.Pos(), token.FUNC, blank)
	if d.Recv != nil {
		p.parameters(d.Recv) // method: print receiver
		p.print(blank)
	}
	p.expr(d.Name)
	p.signature(d.Type.Params, d.Type.Results)
	p.adjBlock(p.distanceFrom(d.Pos()), vtab, d.Body)
}

func (p *printer) decl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.BadDecl:
		p.print(d.Pos(), "BadDecl")
	case *ast.GenDecl:
		p.genDecl(d)
	case *ast.FuncDecl:
		p.funcDecl(d)
	default:
		panic("unreachable")
	}
}

// ----------------------------------------------------------------------------
// Files

func declToken(decl ast.Decl) (tok token.Token) {
	tok = token.ILLEGAL
	switch d := decl.(type) {
	case *ast.GenDecl:
		tok = d.Tok
	case *ast.FuncDecl:
		tok = token.FUNC
	}
	return
}

func (p *printer) declList(list []ast.Decl) {
	tok := token.ILLEGAL
	for _, d := range list {
		prev := tok
		tok = declToken(d)
		// If the declaration token changed (e.g., from CONST to TYPE)
		// or the next declaration has documentation associated with it,
		// print an empty line between top-level declarations.
		// (because p.linebreak is called with the position of d, which
		// is past any documentation, the minimum requirement is satisfied
		// even w/o the extra getDoc(d) nil-check - leave it in case the
		// linebreak logic improves - there's already a TODO).
		if len(p.output) > 0 {
			// only print line break if we are not at the beginning of the output
			// (i.e., we are not printing only a partial program)
			min := 1
			if prev != tok || getDoc(d) != nil {
				min = 2
			}
			p.linebreak(p.lineFor(d.Pos()), min, ignore, false)
		}
		p.decl(d)
	}
}

func (p *printer) file(src *ast.File) {
	p.setComment(src.Doc)
	p.print(src.Pos(), token.PACKAGE, blank)
	p.expr(src.Name)
	p.declList(src.Decls)
	p.print(newline)
}

// ================================================================================
