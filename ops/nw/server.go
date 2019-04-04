// Copyright (C) 2018 rameshvk. All rights reserved.
// Use of this source code is governed by a MIT-style license
// that can be found in the LICENSE file.

package nw

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dotchain/dot/ops"
)

// Handler implements ServerHTTP using the provided store and codecs
// map. If no codecs map is provided, DefaultCodecs is used instead.
type Handler struct {
	ops.Store
	Codecs map[string]Codec
}

// ServeHTTP uses the code to unmarshal a request, apply it and then
// encode back the response
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		ignore(r.Body.Close())
	}()

	ct := r.Header.Get("Content-Type")

	codecs := h.Codecs
	if codecs == nil {
		codecs = DefaultCodecs
	}

	codec := codecs[ct]
	if codec == nil {
		log.Println("Client used an unknown type", ct)
		http.Error(w, "Invalid content-type", 400)
		return
	}

	var req request
	err := codec.Decode(&req, r.Body)
	if err != nil {
		log.Println("Decoding error (see https://github.com/dotchain/dot/wiki/Gob-error)")
		log.Println(err)
		http.Error(w, err.Error(), 400)
		return
	}

	duration := 30 * time.Second
	if req.Duration != 0 {
		duration = req.Duration
	}

	ctx, done := context.WithTimeout(r.Context(), duration)
	defer done()

	var res response
	res.Error = errors.New("unknown error")
	switch req.Name {
	case "Append":
		res.Error = h.Append(ctx, req.Ops)
	case "GetSince":
		res.Ops, res.Error = h.GetSince(ctx, req.Version, req.Limit)
	case "Poll":
		res.Error = h.Poll(ctx, req.Version)
	}

	// do this hack since we can't be sure what error types are possible
	if res.Error != nil {
		if res.Error != ctx.Err() {
			log.Println(req.Name, "failed", res.Error)
		}
		res.Error = strError(res.Error.Error())
	}

	var buf bytes.Buffer
	if err := codec.Encode(&res, &buf); err != nil {
		log.Println("Encoding error (see https://github.com/dotchain/dot/wiki/Gob-error)")
		log.Println(err)

		http.Error(w, err.Error(), 400)
		return
	}
	w.Header().Add("Content-Type", ct)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Println("Unexpected write error", err, req.Name, res)
	}
}

func ignore(err error) {}
