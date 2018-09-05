// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package types

import "github.com/dotchain/dot/changes"

// The A type represents a slice of arbitrary values. It implements
// the changes.Value interface. The actual elements can be nil (unlike
// the regular requirement for values to be non-nil). Nil values are
// treated as if they were changes.Nil
type A []changes.Value

// Slice implements changes.Value.Slice
func (a A) Slice(offset, count int) changes.Value {
	return a[offset : offset+count]
}

// Count returns size of the array
func (a A) Count() int {
	return len(a)
}

// Apply applies the change and returns the updated value
//
// Note: deleting an element via changes.Replace simply replaces it
// with nil.  It does not actually remove the element -- that needs a
// changes.Splice.
func (a A) Apply(c changes.Change) changes.Value {
	switch c := c.(type) {
	case nil:
		return a
	case changes.Replace:
		if c.IsDelete {
			return changes.Nil
		}
		return c.After
	case changes.Splice:
		remove := c.Before.Count()
		after := c.After.(A)
		start, end := c.Offset, c.Offset+remove
		return append(append(a[:start:start], after...), a[end:]...)
	case changes.Move:
		ox, cx, dx := c.Offset, c.Count, c.Distance
		if dx < 0 {
			ox, cx, dx = ox+dx, -dx, cx
		}
		x1, x2, x3 := ox, ox+cx, ox+cx+dx
		return append(append(append(a[:x1:x1], a[x2:x3]...), a[x1:x2]...), a[x3:]...)
	case changes.PathChange:
		if len(c.Path) > 0 {
			idx := c.Path[0].(int)
			clone := append([]changes.Value(nil), a...)
			if clone[idx] == nil {
				clone[idx] = changes.Nil
			}
			clone[idx] = clone[idx].Apply(changes.PathChange{c.Path[1:], c.Change})
			if clone[idx] == changes.Nil {
				clone[idx] = nil
			}

			return A(clone)
		}
		return c.ApplyTo(a)
	case changes.Custom:
		return c.ApplyTo(a)
	}
	panic("Unexpected change on Apply")
}
