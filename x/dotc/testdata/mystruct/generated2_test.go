// Generated.  DO NOT EDIT.
package mystruct

import (
	"reflect"
	"testing"

	"github.com/dotchain/dot/changes"
	"github.com/dotchain/dot/streams"
)

func TestmyStructpStream(t *testing.T) {
	s := streams.New()
	values := valuesFormyStructpStream()
	strong := &myStructpStream{Stream: s, Value: values[0]}

	strong = strong.Update(values[1])
	if !reflect.DeepEqual(strong.Value, values[1]) {
		t.Error("Update did not change value", strong.Value)
	}

	s, c := s.Next()
	if !reflect.DeepEqual(c, changes.Replace{Before: values[0], After: values[1]}) {
		t.Error("Unexpected change", c)
	}

	c = changes.Replace{Before: values[1], After: values[2]}
	s = s.Append(c)
	c = changes.Replace{Before: values[2], After: values[3]}
	s = s.Append(c)
	strong = strong.Latest()

	if !reflect.DeepEqual(strong.Value, values[3]) {
		t.Error("Unexpected value", strong.Value)
	}

	_, c = strong.Next()
	if c != nil {
		t.Error("Unexpected change on stream", c)
	}

	s = s.Append(changes.Replace{Before: values[3], After: changes.Nil})
	if strong, c = strong.Next(); c != nil {
		t.Error("Unexpected change on terminated stream", c)
	}

	s.Append(changes.Replace{Before: changes.Nil, After: values[3]})
	if _, c = strong.Next(); c != nil {
		t.Error("Unexpected change on terminated stream", c)
	}
}

func TestmyStructpStreamboo(t *testing.T) {
	s := streams.New()
	values := valuesFormyStructpStream()
	strong := &myStructpStream{Stream: s, Value: values[0]}
	if !reflect.DeepEqual(strong.Value.boo, strong.boo().Value) {
		t.Error("Substream returned unexpected value", strong.boo().Value)
	}
}
func TestmyStructpStreamstr(t *testing.T) {
	s := streams.New()
	values := valuesFormyStructpStream()
	strong := &myStructpStream{Stream: s, Value: values[0]}
	if !reflect.DeepEqual(strong.Value.str, strong.str().Value) {
		t.Error("Substream returned unexpected value", strong.str().Value)
	}
}
func TestmyStructpStreamStr16(t *testing.T) {
	s := streams.New()
	values := valuesFormyStructpStream()
	strong := &myStructpStream{Stream: s, Value: values[0]}
	if !reflect.DeepEqual(strong.Value.Str16, strong.Str16().Value) {
		t.Error("Substream returned unexpected value", strong.Str16().Value)
	}
}