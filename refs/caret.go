// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package refs

import "github.com/dotchain/dot/changes"

// Caret is a selection into a specific position in an array-like
// object.
//
// This is an immutable type -- none of the methds modify the provided
// path itself.
//
// This only handles the standard set of changes. Custom changes
// should implement a MergeCaret method:
//
//    MergeCaret(caret refs.Caret) (refs.Ref)
//
// Note that this is in addition to the MergePath method which is
// called first to transform the path and then the MergeCaret is
// called  on the updated Caret (based on the path returned by
// MergePath).
type Caret struct {
	Path
	Index int
}

// Merge updates the caret index based on the change.  Note that it
// always returns a nil change as there is no way for a change to
// affect the caret.
func (caret Caret) Merge(c changes.Change) (Ref, changes.Change) {
	px, cx := caret.Path.Merge(c)
	if px == InvalidRef {
		return px, cx
	}
	return caret.updateIndex(px.(Path), caret.Index, cx), nil
}

func (caret Caret) updateIndex(path Path, idx int, cx changes.Change) Ref {
	switch cx := cx.(type) {
	case changes.Replace:
		return InvalidRef
	case changes.Splice:
		return Caret{path, mapIndex(cx, idx)}
	case changes.Move:
		return Caret{path, mapIndex(cx, idx)}
	case changes.PathChange:
		if len(cx.Path) == 0 {
			return caret.updateIndex(path, idx, cx.Change)
		}
	case changes.ChangeSet:
		for _, c := range cx {
			ref := caret.updateIndex(path, idx, c)
			if ref == InvalidRef {
				return ref
			}
			idx = ref.(Caret).Index
		}
	case caretMerger:
		return cx.MergeCaret(Caret{path, idx})
	}
	return Caret{path, idx}
}

func mapIndex(c changes.Change, idx int) int {
	if m, ok := c.(changes.Move); ok {
		switch {
		case idx >= m.Offset+m.Distance && idx < m.Offset:
			idx += m.Count
		case idx >= m.Offset && idx < m.Offset+m.Count:
			idx += m.Distance
		case idx >= m.Offset+m.Count && idx < m.Offset+m.Count+m.Distance:
			idx -= m.Count
		}
		return idx
	}
	cx := c.(changes.Splice)
	if idx >= cx.Offset && idx < cx.Offset+cx.Before.Count() {
		idx = cx.Offset
	}
	if idx >= cx.Offset+cx.Before.Count() {
		idx += cx.After.Count() - cx.Before.Count()
	}
	return idx
}

type caretMerger interface {
	MergeCaret(caret Caret) Ref
}
