// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package dotc

import "io"

// Slice has the type information of a slice for code generation of
// the Apply(), ApplyCollection() and Splice() methods
type Slice struct {
	Recv, Type, ElemType string
	Atomic               bool
}

// RawType returns the non-pointer inner type if Type is a pointer
func (s Slice) RawType() string {
	if s.Pointer() {
		return s.Type[1:]
	}
	return s.Type
}

// Pointer checks if the slice type is a pointer
func (s Slice) Pointer() bool {
	return s.Type[0] == '*'
}

// WrapR
func (s Slice) WrapR(recv string) string {
	if s.Pointer() {
		recv = "(*" + recv + ")"
	}
	return s.Wrap(recv + "[key.(int)]")
}

// Wrap
func (s Slice) Wrap(val string) string {
	switch {
	case s.Atomic || basic[s.ElemType]:
		return "changes.Atomic{" + val + "}"
	case s.ElemType == "string":
		return "types.S16(" + val + ")"
	}
	return val
}

// Unwrap converts "v" to the type of the field
func (s Slice) Unwrap() string {
	if s.Atomic || basic[s.ElemType] {
		return "v.(changes.Atomic).Value.(" + s.ElemType + ")"
	}

	if s.ElemType == "string" {
		return "string(v.(types.S16))"
	}
	return "v.(" + s.ElemType + ")"
}

// GenerateApply generates the code for the changes.Value Apply() method
// and the ApplyCollection() method
func (s Slice) GenerateApply(w io.Writer) error {
	return sliceApply.Execute(w, s)
}

// GenerateSetters generates the code for the field setters
func (s Slice) GenerateSetters(w io.Writer) error {
	return sliceSetters.Execute(w, s)
}
