// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// Package text implements editable text streams
package text

import (
	"github.com/dotchain/dot/changes"
	"github.com/dotchain/dot/refs"
	"github.com/dotchain/dot/x/types"
	"golang.org/x/text/unicode/norm"
)

// Editable implements text editing functionality.  The main state
// maintained by Editable is the actual Text, the current location of
// the cursor and a set of selections that can be maintained with the
// text.
//
// Editable is an immutable type.  All mutations return a
// change.Change and the updated value
//
// There are two positions for each index: left or right. This is
// relevant when considering text that has wrapped around. The
// index in the text where wrapping occurs has two different positions
// on the screen: at the end of the line before wrapping and at the
// start of the line after wrapping.  The top position is considered
// "left" and the bottom line position is considered "right".
//
// There is another consideration: when a remote change causes an
// insertion at exactly the index of the cursor/caret, the caret can
// either be left alone or the caret can be pushed to the right by the
// inserted text.  The "left" position and "right" position match the
// two behaviors (respectively)
type Editable struct {
	Text   string
	Cursor refs.Range
	Refs   map[interface{}]refs.Ref
	Use16  bool

	// SessionID uniquely defines the current "session"
	SessionID interface{}
}

var p = refs.Path{"Value"}

// WithSessionID returns a new stream with an updated SessionID
func (e *Editable) WithSessionID(id interface{}) *Editable {
	result := *e
	result.SessionID = id
	return &result
}

// SetSelection sets the selection range for text.
func (e *Editable) SetSelection(start, end int, left bool) (changes.Change, *Editable) {
	start, end = e.toValueOffset(start), e.toValueOffset(end)
	startx := refs.Caret{p, start, start > end || start == end && left}
	endx := refs.Caret{p, end, start < end || start == end && left}
	l, c := e.toList().UpdateRef(e.SessionID, refs.Range{startx, endx})
	return c, e.fromList(l)

}

// Insert inserts strings at the current cursor position.  If the
// cursor is not collapsed, it collapses the cursor)
func (e *Editable) Insert(s string) (changes.Change, *Editable) {
	offset, before := e.selection()
	after := e.stringToValue(s)
	splice := changes.PathChange{p, changes.Splice{offset, before, after}}
	l := e.toList().Apply(splice).(refs.Container)
	caret := refs.Caret{p, offset + after.Count(), false}
	lx, cx := l.UpdateRef(e.SessionID, refs.Range{caret, caret})
	return changes.ChangeSet{splice, cx}, e.fromList(lx)
}

// Delete deletes the selection. In the case of a collapsed selection,
// it deletes the last character
func (e *Editable) Delete() (changes.Change, *Editable) {
	offset, before := e.selection()
	if offset == 0 && before.Count() == 0 {
		return nil, e
	}

	after := before.Slice(0, 0)
	caret := refs.Caret{p, offset, true}

	if before.Count() == 0 {
		idx := e.fromValueOffset(offset)
		idx -= e.PrevCharWidth(idx)
		caret.Index = e.toValueOffset(idx)
		before = e.stringToValue(e.Text).Slice(caret.Index, offset-caret.Index)
		offset = caret.Index
	}

	splice := changes.PathChange{p, changes.Splice{offset, before, after}}
	l := e.toList()
	lx, cx := l.UpdateRef(e.SessionID, refs.Range{caret, caret})
	lx = lx.Apply(splice).(refs.Container)
	return changes.ChangeSet{cx, splice}, e.fromList(lx)
}

// ArrowLeft implements left arrow key, taking care to properly account
// for unicode sequences.
func (e *Editable) ArrowLeft() (changes.Change, *Editable) {
	idx := e.fromValueOffset(e.cursor().End.Index)
	idx -= e.PrevCharWidth(idx)
	return e.SetSelection(idx, idx, true)
}

// ShiftArrowLeft implements shift left arrow key, taking care to
// properly account for unicode sequences.
func (e *Editable) ShiftArrowLeft() (changes.Change, *Editable) {
	idx := e.fromValueOffset(e.cursor().End.Index)
	idx -= e.PrevCharWidth(idx)
	s := e.fromValueOffset(e.cursor().Start.Index)
	return e.SetSelection(s, idx, true)
}

// ArrowRight implements right arrow key, taking care to properly account
// for unicode sequences.
func (e *Editable) ArrowRight() (changes.Change, *Editable) {
	idx := e.fromValueOffset(e.cursor().End.Index)
	idx += e.NextCharWidth(idx)
	return e.SetSelection(idx, idx, false)
}

// ShiftArrowRight implements shift right arrow key, taking care to
// properly account for unicode sequences.
func (e *Editable) ShiftArrowRight() (changes.Change, *Editable) {
	idx := e.fromValueOffset(e.cursor().End.Index)
	idx += e.NextCharWidth(idx)
	s := e.fromValueOffset(e.cursor().Start.Index)
	return e.SetSelection(s, idx, false)
}

// Copy does not change editable.  It just returns the text currently
// selected.
func (e *Editable) Copy() string {
	_, sel := e.selection()
	return e.valueToString(sel)
}

// Start returns the cursor index. If utf16 is set, it returns the
// offset in UTF16 units. Otherwise in utf8 units
func (e *Editable) Start(utf16 bool) (int, bool) {
	return e.caretToIndex(e.cursor().Start, utf16)
}

// End returns the cursor end.  If utf16 is set, it returns the offset
// in UTF16 units. Otherwise in utf8 units
func (e *Editable) End(utf16 bool) (int, bool) {
	return e.caretToIndex(e.cursor().End, utf16)
}

// StartOf returns the cursor index of the specified session. If utf16
// is set, it returns the offset in UTF16 units. Otherwise in utf8
// units
func (e *Editable) StartOf(sessionID interface{}, utf16 bool) (int, bool) {
	return e.caretToIndex(e.Refs[sessionID].(refs.Range).Start, utf16)
}

// EndOf returns the cursor index of the specified session.  If utf16
// is set, it returns the offset in UTF16 units. Otherwise in utf8 units
func (e *Editable) EndOf(sessionID interface{}, utf16 bool) (int, bool) {
	return e.caretToIndex(e.Refs[sessionID].(refs.Range).End, utf16)
}

func (e *Editable) caretToIndex(caret refs.Caret, utf16 bool) (int, bool) {
	if e.Use16 == utf16 {
		return caret.Index, caret.IsLeft
	}
	if utf16 {
		return types.S16(e.Text).ToUTF16(caret.Index), caret.IsLeft
	}
	return e.fromValueOffset(caret.Index), caret.IsLeft
}

// Value just returns the inner Text.  This is mainly there to make it
// easier to use this function from Javascript-land
func (e *Editable) Value() string {
	return e.Text
}

// Paste is like insert except it keeps the cursor around the pasted
// string.
func (e *Editable) Paste(s string) (changes.Change, *Editable) {
	offset, before := e.selection()
	after := e.stringToValue(s)
	splice := changes.PathChange{p, changes.Splice{offset, before, after}}
	l := e.toList().Apply(splice).(refs.Container)
	start := refs.Caret{p, offset, after.Count() == 0}
	end := refs.Caret{p, offset + after.Count(), true}
	lx, cx := l.UpdateRef(e.SessionID, refs.Range{start, end})
	return changes.ChangeSet{splice, cx}, e.fromList(lx)
}

// Apply implements the changes.Value interface
func (e *Editable) Apply(c changes.Change) changes.Value {
	result := e.toList().Apply(c)
	l, ok := result.(refs.Container)
	if !ok {
		return result
	}

	return e.fromList(l)
}

func (e *Editable) stringToValue(s string) changes.Collection {
	if e.Use16 {
		return types.S16(s)
	}
	return types.S8(s)
}

func (e *Editable) valueToString(v changes.Value) string {
	if e.Use16 {
		return string(v.(types.S16))
	}
	return string(v.(types.S8))
}

func (e *Editable) cursor() refs.Range {
	c := e.Cursor
	c.Start.Path = p
	c.End.Path = p
	return c
}

func (e *Editable) toValueOffset(idx int) int {
	if e.Use16 {
		return types.S16(e.Text).ToUTF16(idx)
	}
	// validate that the offset works
	_ = e.Text[idx:]
	return idx
}

func (e *Editable) fromValueOffset(idx int) int {
	if e.Use16 {
		return types.S16(e.Text).FromUTF16(idx)
	}
	// validate that the offset works
	_ = e.Text[idx:]
	return idx
}

func (e *Editable) toList() refs.Container {
	return refs.NewContainer(e.stringToValue(e.Text), e.Refs)
}

func (e *Editable) fromList(l refs.Container) *Editable {
	text := e.valueToString(l.Value)
	cursor, ok := l.GetRef(e.SessionID).(refs.Range)
	if !ok {
		cursor = refs.Range{refs.Caret{Path: p}, refs.Caret{Path: p}}
	}
	return &Editable{text, cursor, l.Refs(), e.Use16, e.SessionID}
}

func (e *Editable) selection() (int, changes.Collection) {
	c := e.cursor()
	v := e.stringToValue(e.Text)
	start, end := c.Start.Index, c.End.Index
	diff := end - start
	if start > end {
		start, end = end, start
		diff = end - start
	}
	return start, v.Slice(start, diff)
}

// NextCharWidth returns the width of a user-perceived character.  This
// takes care of combining characters and such.
func (e *Editable) NextCharWidth(idx int) int {
	return norm.NFC.NextBoundaryInString(e.Text[idx:], true)
}

// PrevCharWidth returns the width of a user-perceived character
// before the provided index.  This takes care of combining characters
// and such.
func (e *Editable) PrevCharWidth(idx int) int {
	text := []byte(e.Text)[:idx]

	offset := norm.NFC.LastBoundary(text)
	if offset < 0 {
		return 0
	}

	if offset < idx || idx == 0 {
		return idx - offset
	}

	// NFC.LastBoundary is quite buggy in some cases.
	// See: https://github.com/golang/go/issues/9055
	// The work around is to brute force it in those cases
	idx = len(text) - 100
	if idx < 0 {
		idx = 0
	}
	w := len(text) - idx
	for w > 1 && norm.NFC.NextBoundary(text[idx:], true) != w {
		idx++
		w--
	}
	return w
}
