// Copyright (C) 2018 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package nw_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/dotchain/dot/changes"
	"github.com/dotchain/dot/ops"
	"github.com/dotchain/dot/ops/nw"

	"github.com/dotchain/dot/test/testops"
)

func Example() {
	store := ops.Polled(testops.MemStore(nil))
	defer store.Close()
	handler := &nw.Handler{Store: store}
	srv := httptest.NewServer(handler)
	defer srv.Close()

	c := nw.Client{URL: srv.URL, Client: srv.Client()}
	defer c.Close()

	op1 := ops.Operation{OpID: "ID1", ParentID: "", VerID: 100, BasisID: -1}
	op2 := ops.Operation{OpID: "ID2", ParentID: "ID1", VerID: 100, BasisID: -1, Change: changes.ChangeSet{changes.Move{Offset: 1, Count: 2, Distance: 3}}}

	ctx := getContext()
	if err := c.Append(ctx, []ops.Op{op1}); err != nil {
		fmt.Println("Append1", err)
		return
	}
	if err := c.Append(ctx, []ops.Op{op2}); err != nil {
		fmt.Println("Append2", err)
		return
	}

	ops, err := c.GetSince(ctx, 0, 100)
	fmt.Println("Ops", ops, err)

	// Output:
	// Ops [{ID1  0 -1 <nil>} {ID2 ID1 1 -1 [{1 2 3}]}] <nil>
}
