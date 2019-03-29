// Code generated by github.com/tvastar/test/cmd/testmd/testmd.go. DO NOT EDIT.

package example

import (
	"encoding/gob"
	"math/rand"
	"net/http"

	"github.com/dotchain/dot/ops"
	"github.com/dotchain/dot/ops/bolt"
	"github.com/dotchain/dot/ops/nw"
)

func Server() {
	// import net/http
	// import github.com/dotchain/dot/ops/nw
	// import github.com/dotchain/dot/ops/bolt

	// uses a local-file backed bolt DB backend
	store, _ := bolt.New("file.bolt", "instance", nil)
	defer store.Close()
	http.Handle("/api/", &nw.Handler{Store: store})
	http.ListenAndServe(":8080", nil)
}

// Todo tracks a single todo item
type Todo struct {
	Complete    bool
	Description string
}

// TodoList tracks a collection of todo items
type TodoList []Todo

// import encoding/gob

func init() {
	gob.Register(Todo{})
	gob.Register(TodoList{})
}
func Toggle(t *TodoListStream, index int) {
	// TodoListStream.Item() is implemented in the generated
	// code and returns *TodoStream
	itemStream := t.Item(index)

	// Complete() is also implemented in the generated code.
	completeStream := itemStream.Complete()

	// Update() here refers to streams.Bool.Update
	completeStream.Update(!completeStream.Value)
}
func SpliceDescription(t *TodoListStream, index, offset, count int, replacement string) {
	// TodoListStream.Item() is implemented in the generated
	// code and returns *TodoStream
	itemStream := t.Item(index)

	// Description() is also implemented in the generated code.
	descStream := itemStream.Description()

	// Splice() here refers to streams.S16.Splice
	descStream.Splice(offset, count, replacement)
}
func AddTodo(t *TodoListStream, todo Todo) {
	t.Splice(len(t.Value), 0, todo)
}

// import github.com/dotchain/dot/ops/nw
// import github.com/dotchain/dot/ops
// import math/rand

func Client(stop chan struct{}, url string, render func(*TodoListStream)) {
	version, pending, todos := SavedSession()

	store := &nw.Client{URL: url}
	defer store.Close()
	client := ops.NewConnector(version, pending, ops.Transformed(store), rand.Float64)
	stream := &TodoListStream{Stream: client.Stream, Value: todos}

	// start the network processing
	client.Connect()

	// save session before shutdown
	defer func() {
		SaveSession(client.Version, client.Pending, stream.Latest().Value)
	}()
	defer client.Disconnect()

	client.Stream.Nextf("key", func() {
		stream = stream.Latest()
		render(stream)
	})
	render(stream)
	defer func() {
		client.Stream.Nextf("key", nil)
	}()

	<-stop
}

func SaveSession(version int, pending []ops.Op, todos TodoList) {
	// this is not yet implemented. if it were, then
	// this value should be persisted locally and returned
	// by the call to savedSession
}

func SavedSession() (version int, pending []ops.Op, todos TodoList) {
	// this is not yet implemented. return default values
	return -1, nil, nil
}