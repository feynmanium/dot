// Copyright (C) 2017 Ramesh Vyaghrapuri. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package dot

// ClientLog is a helper struct that provides the functionality
// needed by clients to deal with operations from other clients
// that get merged into the journal.  It maintains the state
// needed to calculate the "compensating" operations that a
// client can apply to get to the same converged state that
// would result if its inflight operations were merged into the
// journal.  Note that this is a moving target -- as more operations
// get added to the journal and the local client keeps adding
// more operations of its own, the compensating operations need
// to track both and yield a final converged state that mirrors
// what silent observer would end up with if it only tracked
// the server log.
//
// Please read https://github.com/dotchain/dot/docs/IntroductionToOperationalTransforms.md
// for a detailed description of how reconciliation works.
//
// The initialization of the client log involves a few different
// cases.
//
// 1. Client starts from scratch, no model at all but might have
// operations from its previous session on the device that were
// in flight (maybe in the journal, may not)
//
// In this case, the client should use #BootstrapClientLog to
// bootstrap its model.
//
// 2. Client has restarted a session with a cached model at a
// particular basis and potentially some client operations that
// were in flight before.
//
// In this case, the client should use #ReconnectClientLog to continue
// the reconciliation process
//
// Please see
// https://github.com/dotchain/site/blob/master/Protocol.md
// for a better understanding of the use of the ParentID and
// BasisID when sending them to the server.  #AppendClientOperation
// expects these to be properly setup with BasisID being the
// last value in server log when the client applied the operation
// and ParentID being the last client operation applied in the
// current session (or carried over from a previous session)
type ClientLog struct {
	Transformer

	// the following two numbers are 1+ index in server log
	// with the 1 being there because of making it easy to
	// initialize the log with zeroes.

	// 1 + index of basis of last known client operation
	ClientIndex int
	// 1 + index of last known operation from server log
	// that has been factored into the client log so far
	ServerIndex int

	// Rebased maintains the rebased client operations that
	// have yet to appear in the server log
	Rebased []Operation

	// MergeChain is the sequence of operations to apply after
	// the last rebased operation to get the model into a
	// converged state.  This is empty if rebased is empty
	MergeChain []Operation
}

// Reconcile takes a server log and if there are any operations
// there that have not been added to the client log, it updates
// the client log.  It returns the set of compensating operations
// to apply to the client model to get to the converged state.
func (c *ClientLog) Reconcile(l *Log) ([]Operation, error) {
	var ok bool

	if c.ServerIndex+1 <= l.MinIndex {
		return nil, ErrLogNeedsBackfilling
	}

	rebased, merge := c.Rebased, []Operation{}
	serverIndex := c.ServerIndex
	for _, op := range l.Rebased[c.ServerIndex:] {
		serverIndex++
		if len(rebased) > 0 && rebased[0].ID == op.ID {
			rebased = rebased[1:]
			continue
		}
		var m []Operation
		rebased, m, ok = c.TryMergeOperations([]Operation{op}, rebased)
		if !ok {
			return nil, ErrInvalidOperation
		}
		merge = append(merge, m...)
	}

	c.ServerIndex = serverIndex
	if len(rebased) > 0 {
		c.Rebased = append([]Operation{}, rebased...)
		c.MergeChain = append(c.MergeChain, merge...)
	} else {
		c.Rebased = nil
		c.MergeChain = nil
	}
	return merge, nil
}

// AppendClientOperation appends a client operation to the client log.
// It can be used to initialize a client log or to append to a client log
// that has already appended a few client operations before.
//
// It returns an error if the server log needs backfilling. The returned
// set of compensating operations can be used by the client to update
// its state to factor in the effect of any unaccounted ops in the log.
func (c *ClientLog) AppendClientOperation(l *Log, op Operation) ([]Operation, error) {
	var ok bool

	if len(c.Rebased) == 0 {
		return c.initializeFromOperation(l, op)
	}

	lastBasisID := c.Rebased[len(c.Rebased)-1].BasisID()
	mergeChain := c.MergeChain
	clientIndex := c.ClientIndex
	if basisID := op.BasisID(); basisID != lastBasisID {
		index := -1
		for kk, m := range mergeChain {
			if m.ID == basisID {
				clientIndex = l.IDToIndexMap[m.ID] + 1
				index = kk
				break
			}
		}
		if index < 0 {
			return nil, ErrMissingParentOrBasis
		}
		mergeChain = append([]Operation{}, mergeChain[index+1:]...)
	}

	rebased, merge, ok := c.TryMergeOperations(mergeChain, []Operation{op})
	if !ok {
		return nil, ErrInvalidOperation
	}

	c.ClientIndex = clientIndex
	c.MergeChain = append([]Operation{}, merge...)
	c.Rebased = append(c.Rebased, rebased...)

	if c.ServerIndex < len(l.Rebased) {
		return c.Reconcile(l)
	}

	return merge, nil
}

// initializeFromJournal initializes a client log from an operation
// that is present in the journal.  The client can fully reconstitute
// its state by applying all rebased operations in the server log
// until the basis of this operation followed by this operation and
// the return value from this function.
func (c *ClientLog) initializeFromJournal(l *Log, id string) ([]Operation, error) {
	index := l.IDToIndexMap[id]

	if index < l.MinIndex {
		return nil, ErrLogNeedsBackfilling
	}

	basisID := l.Rebased[index].BasisID()
	if basisID == "" {
		c.ClientIndex = 0
	} else {
		c.ClientIndex = l.IDToIndexMap[basisID] + 1
	}
	c.ServerIndex = len(l.Rebased)
	c.Rebased = nil
	c.MergeChain = nil
	return c.joinOperation(l.MergeChains[index], l.Rebased[index+1:]), nil
}

// initializeFromOperation initializes a client log from a client
// operation that may or may not exist yet on the server rebased log.
//
// A client can reconstruct a convergent model by applying the
// operations from the rebased server log up to the basis of the
// provided operation followed by the client operation and then
// followed by the return value from this function
func (c *ClientLog) initializeFromOperation(l *Log, op Operation) ([]Operation, error) {
	if _, ok := l.IDToIndexMap[op.ID]; ok {
		return c.initializeFromJournal(l, op.ID)
	}

	basisIndex, ok := l.IDToIndexMap[op.BasisID()]
	if !ok && op.BasisID() != "" {
		return nil, ErrMissingParentOrBasis
	}

	if op.BasisID() != "" && basisIndex < l.MinIndex || op.BasisID() == "" && l.MinIndex > 0 {
		return nil, ErrLogNeedsBackfilling
	}

	parentIndex, ok := l.IDToIndexMap[op.ParentID()]
	if op.ParentID() != "" && !ok {
		return nil, ErrMissingParentOrBasis
	}

	if op.BasisID() == "" {
		basisIndex = -1
	}

	return c.initialize(l, op, basisIndex, parentIndex)
}

// initialize a client log with a new operation with the provided
// basis and parent indices (which have been error checked already)
func (c *ClientLog) initialize(l *Log, op Operation, basisIndex, parentIndex int) ([]Operation, error) {
	var mergeChain []Operation
	if op.ParentID() == "" || parentIndex <= basisIndex {
		mergeChain = l.Rebased[basisIndex+1:]
	} else {
		mergeChain = c.joinOperation(l.MergeChains[parentIndex], l.Rebased[parentIndex+1:])
		mergeChain = l.TrimMergeChain(mergeChain, op.BasisID())
	}

	r, m, ok := c.TryMergeOperations(mergeChain, []Operation{op})
	if !ok {
		return nil, ErrInvalidOperation
	}

	c.Rebased, c.MergeChain = r, m
	c.MergeChain = append([]Operation{}, c.MergeChain...)
	c.ClientIndex = basisIndex + 1
	c.ServerIndex = len(l.Rebased)

	return c.MergeChain, nil
}

// BootstrapClientLog creates a new client log for a client that does
// not have a model.
//
// Errors: It returns ErrMissingParentOrBasis if the log
// has not advanced enough for the clientOps.  It can return
// ErrLogNeedsBackfilling if the log is not backfilled enough for the
// operation to complete.
//
// It returns the client log and a pair of operation collections. The
// first operation collection is the set of rebased server operations
// a client can apply to get to a good state and the second operations
// collection is the set of rebased client operations which is meant
// to be applied on top of the server rebased.
func BootstrapClientLog(l *Log, clientOps []Operation) (*ClientLog, []Operation, []Operation, error) {
	clog := &ClientLog{Transformer: l.Transformer}
	if _, err := clog.Reconcile(l); err != nil {
		return nil, nil, nil, err
	}

	for _, op := range clientOps {
		if _, err := clog.AppendClientOperation(l, op); err != nil {
			return nil, nil, nil, err
		}
	}

	rebased := append([]Operation{}, l.Rebased...)
	clientRebased := append([]Operation{}, clog.Rebased...)
	return clog, rebased, clientRebased, nil
}

// ReconnectClientLog creates a new client log for a client that has
// an existing model (with the provided parentID and basisID). Note
// that if client operations are provided, the parentID will be
// ignored and the last op in that list will be used.
//
// Errors: It returns ErrMissingParentOrBasis if the log
// has not advanced enough for the clientOps.  It can return
// ErrLogNeedsBackfilling if the log is not backfilled enough for the
// operation to complete.
//
// It also returns a set of operations that the client can apply to
// get it back to a mainline state
func ReconnectClientLog(l *Log, clientOps []Operation, basisID, parentID string) (*ClientLog, []Operation, error) {
	clog := &ClientLog{Transformer: l.Transformer}

	if len(clientOps) > 0 {
		parentID = clientOps[len(clientOps)-1].ID
	}

	if _, ok := l.IDToIndexMap[basisID]; !ok {
		return nil, nil, ErrMissingParentOrBasis
	}
	clog.ClientIndex = l.IDToIndexMap[basisID] + 1
	if _, err := clog.Reconcile(l); err != nil {
		return nil, nil, err
	}
	for _, op := range clientOps {
		if _, err := clog.AppendClientOperation(l, op); err != nil {
			return nil, nil, err
		}
	}

	merge := clog.MergeChain
	if merge == nil {
		merge = l.Rebased[l.MinIndex:]
	}
	merge = l.TrimMergeChain(l.TrimMergeChain(merge, basisID), parentID)
	return clog, merge, nil
}
