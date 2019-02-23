// This file is generated by:
//    github.com/dotchain/dot/ux/templates/cache.template
//
// Copyright (C) 2019 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package todo

import "github.com/dotchain/dot/ux/core"

// TasksViewCache holds a cache of TasksView controls.
//
// Controls that have manage a bunch of TasksView controls
// should maintain a cache created like so:
//
//     cache := &TasksViewCache{}
//
// When updating, the cache can be used to reuse controls:
//
//     cache.Begin()
//     defer cache.End()
//
//     ... for each TasksView control needed do:
//     cache.Get(key, styles, done, notDone, tasks)
//
// This allows the cache to reuse the control if the key exists.
// Otherwise a new control is created via NewTasksView(styles, done, notDone, tasks)
//
// When a control is reused, it is also automatically updated.
type TasksViewCache struct {
	old, current map[interface{}]*TasksView
}

// Begin should be called before the start of a round
func (c *TasksViewCache) Begin() {
	c.old = c.current
	c.current = map[interface{}]*TasksView{}
}

// End should be called at the end of a round
func (c *TasksViewCache) End() {
	// if components had a Close() method all the old left-over items
	// can be cleaned up via that call
	c.old = nil
}

// TryGet fetches a TasksView from the cache (updating it)
// or creates a new TasksView
//
// It returns the TasksView but also whether the control existed.
// This can be used to conditionally setup listeners.
func (c *TasksViewCache) TryGet(key interface{}, styles core.Styles, done bool, notDone bool, tasks Tasks) (*TasksView, bool) {
	exists := false
	if item, ok := c.old[key]; !ok {
		c.current[key] = NewTasksView(styles, done, notDone, tasks)
	} else {
		delete(c.old, key)
		item.Update(styles, done, notDone, tasks)
		c.current[key] = item
		exists = true
	}

	return c.current[key], exists
}

// Item fetches the item at the specific key
func (c *TasksViewCache) Item(key interface{}) *TasksView {
	return c.current[key]
}

// Get fetches a TasksView from the cache (updating it)
// or creates a new TasksView
//
// Use TryGet to also fetch whether the control from last round was reused
func (c *TasksViewCache) Get(key interface{}, styles core.Styles, done bool, notDone bool, tasks Tasks) *TasksView {
	v, _ := c.TryGet(key, styles, done, notDone, tasks)
	return v
}