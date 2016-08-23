// Copyright 2016 The Go and Gro Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tutor provides the interface for tutorials in Gro.
//
package tutor

import (
)

//================================================================================

// A Tutorial is a tutorial available.
type Tutorial struct {
	// Name is the name of the tutorial.
	Name string

	// Short is the short description shown in the 'gro tutorial' output.
	Short string

	// Long is the long message shown in the 'gro tutorial <this-tutorial>' output.
	Long string

	// Pages is a slice of the pages in the tutorial.
	Pages []*Page
}

//================================================================================

// A Page is a page in a tutorial.
type Page struct {
	// Code is the code samples in the tutorial.
	Code []string

	// English is the English version of the tutorial page.
	English []string

	// Chinese is the Chinese version of the tutorial page.
	Chinese []string

	// Japanese is the Japanese version of the tutorial page.
	Japanese []string
}

//================================================================================

func (p Page) EnglishPage () (s string) {
	for i, v:= range p.English {
		s = s + v
		if i < len(p.Code) {
			s = s + p.Code[i]
		}
	}
	return
}

//================================================================================

