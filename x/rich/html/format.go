// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package html

import (
	"strings"

	"github.com/dotchain/dot/x/rich"
)

// Formatter incrementally formats a text segment into html
type Formatter interface {
	Open(b *strings.Builder, last, current rich.Attrs, text string)
	Close(b *strings.Builder, last, current rich.Attrs, text string)
}

// Format formats rich text into html
func Format(t rich.Text, f Formatter) string {
	if f == nil {
		f = DefaultFormatter
	}
	var b strings.Builder
	last := rich.Attrs{}
	for _, x := range t {
		f.Close(&b, last, x.Attrs, x.Text)
		f.Open(&b, last, x.Attrs, x.Text)
		last = x.Attrs
	}
	if !last.Equal(rich.Attrs{}) {
		f.Close(&b, last, rich.Attrs{}, "")
	}
	return b.String()
}

// DefaultFormatter formats standard styles such as plain text string,
// bold and italics.
var DefaultFormatter = simpleFmt{
	[]string{"FontStyle", "FontWeight"},
	textFmt{},
}