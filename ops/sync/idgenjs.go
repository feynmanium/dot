// Copyright (C) 2018 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

// +build js,!jsreflect

package sync

import (
	"encoding/base64"

	"github.com/gopherjs/gopherjs/js"
)

// newID returns a unique ID using crypto
func (s *session) newID() (interface{}, error) {
	crypto := js.Global.Get("crypto")
	array := js.Global.Get("Uint8Array").New(32)
	crypto.Call("getRandomValues", array)
	return base64.StdEncoding.EncodeToString(array.Interface().([]byte)), nil
}
