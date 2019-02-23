// This file is generated by:
//    github.com/dotchain/dot/ux/templates/streams.template
//
// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package todo

import "github.com/dotchain/dot/changes"
import "github.com/dotchain/dot/ux/streams"

// TaskStream holds a Task value and tracks changes to it.
//
// Changes can be listened to via the embedded Notifier.  Actual
// change values are tracked in a linked-list using the Next field.
type TaskStream struct {
	// Notifier provides On/Off/Notify support.
	*streams.Notifier

	// Value represents the current value. Use Latest() to get the
	// latest value.
	Value Task

	// Change represents the chagne that results in an updated
	// value. The updated value can be identified via .Next.
	Change changes.Change

	// Next is the next value in the sequence.
	Next *TaskStream
}

// NewTaskStream creates a new Task stream
func NewTaskStream(v Task) *TaskStream {
	return &TaskStream{&streams.Notifier{}, v, nil, nil}
}

// Update updates the stream with a new value and returns the
// latest value.  To notify listeners, an explicit call to Notify
// is required.
func (s *TaskStream) Update(c changes.Change, value Task) *TaskStream {
	if c == nil {
		c = changes.Replace{changes.Nil, changes.Atomic{value}}
	}
	if s.Next != nil {
		// This version does not merge results. The Streams
		// based ReverseMerge() algorithm can be used here but
		// that requires a changes.Value interface
		// implementation in BaseType
		panic("Unexpected update on stale version")
	}

	s.Next = &TaskStream{s.Notifier, value, c, nil}
	return s.Next
}

// Latest returns the latest value in the current stream
func (s *TaskStream) Latest() *TaskStream {
	for s.Next != nil {
		s = s.Next
	}
	return s
}
