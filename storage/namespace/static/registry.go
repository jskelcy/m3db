// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package namespace

import (
	"github.com/m3db/m3db/storage/namespace"
	"github.com/m3db/m3x/watch"
)

type staticInit struct {
	metadatas []namespace.Metadata
}

// NewInitializer returns a new static registry initializer
func NewInitializer(metadatas []namespace.Metadata) namespace.Initializer {
	return &staticInit{
		metadatas: metadatas,
	}
}

func (i *staticInit) Init() (namespace.Registry, error) {
	return newRegistry(i.metadatas)
}

type staticReg struct {
	xwatch.Watchable
}

func newRegistry(metadatas []namespace.Metadata) (namespace.Registry, error) {
	m, err := namespace.NewMap(metadatas)
	if err != nil {
		return nil, err
	}
	reg := &staticReg{xwatch.NewWatchable()}
	if err := reg.Update(m); err != nil {
		return nil, err
	}
	return reg, nil
}

func (r *staticReg) Watch() (namespace.Watch, error) {
	_, w, err := r.Watchable.Watch()
	if err != nil {
		return nil, err
	}
	return namespace.NewWatch(w), nil
}

func (r *staticReg) Close() error {
	r.Watchable.Close()
	return nil
}
