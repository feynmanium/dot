// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package rich_test

import (
	"reflect"
	"testing"

	"github.com/dotchain/dot/changes"
	"github.com/dotchain/dot/x/rich"
	"github.com/dotchain/dot/x/rich/html"
)

func TestAttrs(t *testing.T) {
	a1 := rich.Attrs{"FontWeight": html.FontBold}
	a2 := rich.Attrs{"FontWeight": html.FontThin}
	if !a1.Equal(a1) || a2.Equal(a1) {
		t.Error("Unexpected equality failure")
	}

	c := changes.PathChange{
		Path: []interface{}{"FontWeight"},
		Change: changes.Replace{
			Before: html.FontBold,
			After:  html.FontThin,
		},
	}
	if x := a1.Apply(nil, c); !reflect.DeepEqual(x, a2) {
		t.Error("Apply change failed", x)
	}
}
