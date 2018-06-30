// Copyright (C) 2018 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package vc

import (
	"github.com/dotchain/dot"
	"strconv"
)

// every container implements the simple parent interface
type parent interface {
	// Bubble is called on a parent with the basis of both the
	// previous change and the current change.  The return value
	// is not used  right now but will be used at the root level
	// for tracking changes.  It is not useful at the non-root
	// level
	Bubble(prev, now *basis, changes []dot.Change)

	// Latest is called on a parent with the path to the node in
	// the current tree. Latest returns the value of the current
	// node in the latest tree as well as the corresponding parent
	// interface. It also returns the updated path and the actual
	// basis of that new change
	Latest(path *dot.RefPath, b *basis) (interface{}, parent, []string, *basis)

	// MapPath maps the provided path at the provided basis to the
	// new basis.   The new basis must be one of those retured by
	// Latest.
	MapPath(path *dot.RefPath, from, to *basis) []string
}

// dictitem is a parent of a dictionary entry. it tracks the key of the
// child as well as the actual dictionary container to proxy calls to
type dictitem struct {
	key  string
	dict parent
}

// Bubble updates the list of changes by prepending the key to the
// path and calling the dictionary with that path. It mutates the
// path of the changes but this is ok because version.Update creates a
// copy before calling Bubble
func (item *dictitem) Bubble(prev, now *basis, changes []dot.Change) {
	for key, c := range changes {
		changes[key].Path = append([]string{item.key}, c.Path...)
	}
	item.dict.Bubble(prev, now, changes)
}

// Latest updates the path to prepend the key and then fetches the
// latest dict + updated path and updated parent for the dict. It uses
// the first entry of the path to find the updated key (which should
// not really change for dictionary paths..) and uses that to traverse
// the new dictionary.
func (item *dictitem) Latest(path *dot.RefPath, b *basis) (interface{}, parent, []string, *basis) {
	path = path.Prepend(item.key, nil)
	containerValue, container, newPath, b := item.dict.Latest(path, b)
	if container == nil {
		// path was invalidated by a later change
		return nil, nil, nil, nil
	}

	key, rest := newPath[0], newPath[1:]
	child := utils.C.Get(containerValue).Get(key)
	return child, &dictitem{key: key, dict: container}, rest, b
}

// MapPath tacks on the key to the path and removes it right after
func (item *dictitem) MapPath(path *dot.RefPath, from, to *basis) []string {
	return item.dict.MapPath(path.Prepend(item.key, nil), from, to)[1:]
}

// arrayitem is a parent of an array entry.  It tracks the index of
// the child as well as the actual array container to proxy calls to.
type arrayitem struct {
	index int
	key   string
	array parent
}

// Bubble updates the list of changes by prepending the index to the
// path and calling the array with that path.  It mutates the path of
// changes but this is ok because verison.Update creates a copy before
// calling Bubble
func (item *arrayitem) Bubble(prev, now *basis, changes []dot.Change) {
	for kk, c := range changes {
		changes[kk].Path = append([]string{item.key}, c.Path...)
	}
	item.array.Bubble(prev, now, changes)
}

// Latest updates the path to preend the index and then fetches the
// latest value of the array + updated path.   It uses the updated
// path to figure out the new index and uses that to find the actually
// array element value to return.
func (item *arrayitem) Latest(path *dot.RefPath, b *basis) (interface{}, parent, []string, *basis) {
	path = path.Prepend("", &dot.RefIndex{Index: item.index})
	containerValue, container, newPath, b := item.array.Latest(path, b)
	if container == nil {
		// path was invalidated by later changes
		return nil, nil, nil, nil
	}
	key, rest := newPath[0], newPath[1:]
	index, _ := strconv.Atoi(key)
	child := utils.C.Get(containerValue).Get(key)
	return child, &arrayitem{key: key, index: index, array: container}, rest, b
}

// MapPath tacks on the index to the path and removes it right after
func (item *arrayitem) MapPath(path *dot.RefPath, from, to *basis) []string {
	index := &dot.RefIndex{Index: item.index}
	return item.array.MapPath(path.Prepend("", index), from, to)[1:]
}
