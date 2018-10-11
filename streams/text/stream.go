// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package text

import (
	"github.com/dotchain/dot/changes"
	"github.com/dotchain/dot/refs"
	"github.com/dotchain/dot/streams"
)

// StreamFromString constructs a new text stream
func StreamFromString(initialText string, use16 bool) *Stream {
	return &Stream{&Editable{Text: initialText}, streams.New()}
}

// Stream implements the streams.Stream interface on top of Editable.
//
// Stream is an immutable type.  All mutations return an new instance
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
type Stream struct {
	E *Editable
	S streams.Stream
}

// Append implements streams.Stream:Append
func (s *Stream) Append(c changes.Change) streams.Stream {
	v := s.E.Apply(c)
	sx := s.S.Append(c)
	if e, ok := v.(*Editable); ok {
		return &Stream{e, sx}
	}
	return &streams.ValueStream{v, sx}
}

// ReverseAppend implements streams.Stream:ReverseAppend
func (s *Stream) ReverseAppend(c changes.Change) streams.Stream {
	v := s.E.Apply(c)
	sx := s.S.ReverseAppend(c)
	if e, ok := v.(*Editable); ok {
		return &Stream{e, sx}
	}
	return &streams.ValueStream{v, sx}
}

// Scheduler implements streams.Stream:Scheduler
func (s *Stream) Scheduler() streams.Scheduler {
	return s.S.Scheduler()
}

// WithScheduler implements streams.Stream:WithScheduler
func (s *Stream) WithScheduler(sch streams.Scheduler) streams.Stream {
	return &Stream{s.E, s.S.WithScheduler(sch)}
}

// Next implements streams.Stream.Next
func (s *Stream) Next() (changes.Change, streams.Stream) {
	return s.mapChangeValue(s.S.Next())
}

// Nextf implements streams.Stream.Nextf
func (s *Stream) Nextf(key interface{}, fn func(c changes.Change, str streams.Stream)) {
	if fn == nil {
		s.S.Nextf(key, fn)
	} else {
		var last streams.Stream = s
		s.S.Nextf(key, func(c changes.Change, str streams.Stream) {
			if l, ok := last.(*Stream); ok {
				c, str = l.mapChangeValue(c, str)
				last = str
			}
			fn(c, str)
		})
	}
}

// SetSelection sets the selection range for text
func (s *Stream) SetSelection(start, end int, left bool) *Stream {
	c, e := s.E.SetSelection(start, end, left)
	return &Stream{e, s.S.Append(c)}
}

// Paste inserts the provided string at current cursor (deleting any
// text that might be selected).  It leaves the current text selected.
func (s *Stream) Paste(str string) *Stream {
	c, e := s.E.Paste(str)
	return &Stream{e, s.S.Append(c)}
}

// Insert inserts a string
func (s *Stream) Insert(str string) *Stream {
	c, e := s.E.Insert(str)
	return &Stream{e, s.S.Append(c)}
}

// Delete deletes the current selection or the last caret before the
// caret.
func (s *Stream) Delete() *Stream {
	c, e := s.E.Delete()
	return &Stream{e, s.S.Append(c)}
}

// WithoutOwnCursor returns a stream that can be used to sync with
// remote clients. The local stream contains changes pertaining to the
// local cursor that is not meant to be shared across to remote
// clients.
func (s *Stream) WithoutOwnCursor() streams.Stream {
	var filter func(changes.Change) changes.Change
	filter = func(c changes.Change) changes.Change {
		switch c := c.(type) {
		case refs.Update:
			if c.Key == own {
				return nil
			}
		case changes.PathChange:
			if len(c.Path) == 0 {
				return filter(c.Change)
			}
		case changes.ChangeSet:
			result := changes.ChangeSet(nil)
			for _, cx := range c {
				if x := filter(cx); x != nil {
					result = append(result, x)
				}
			}
			return result
		}
		return c
	}

	return filterChange{filter, s}
}

func (s Stream) mapChangeValue(c changes.Change, str streams.Stream) (changes.Change, streams.Stream) {
	if str == nil {
		return c, str
	}

	v := s.E.Apply(c)
	if e, ok := v.(*Editable); ok {
		return c, &Stream{e, str}
	}
	return c, &streams.ValueStream{v, str}
}

type filterChange struct {
	filter func(c changes.Change) changes.Change
	base   streams.Stream
}

func (f filterChange) Append(c changes.Change) streams.Stream {
	return filterChange{f.filter, f.base.Append(c)}
}

func (f filterChange) ReverseAppend(c changes.Change) streams.Stream {
	return filterChange{f.filter, f.base.ReverseAppend(c)}
}

func (f filterChange) Scheduler() streams.Scheduler {
	return f.base.Scheduler()
}

func (f filterChange) WithScheduler(sch streams.Scheduler) streams.Stream {
	return filterChange{f.filter, f.base.WithScheduler(sch)}
}

func (f filterChange) Next() (changes.Change, streams.Stream) {
	return f.mapChange(f.base.Next())
}

func (f filterChange) Nextf(key interface{}, fn func(c changes.Change, s streams.Stream)) {
	f.base.Nextf(key, f.mapFn(fn))
}

func (f filterChange) mapChange(c changes.Change, s streams.Stream) (changes.Change, streams.Stream) {
	if s != nil {
		s = filterChange{f.filter, s}
	}
	return f.filter(c), s
}

func (f filterChange) mapFn(fn func(changes.Change, streams.Stream)) func(changes.Change, streams.Stream) {
	if fn != nil {
		old := fn
		fn = func(c changes.Change, s streams.Stream) {
			old(f.mapChange(c, s))
		}
	}
	return fn
}
