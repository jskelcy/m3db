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
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/m3db/m3db/ts"
)

func TestRegistryEmptyNamespaceRegistry(t *testing.T) {
	emptyRegistry := NewRegistry(nil)
	require.Empty(t, emptyRegistry.IDs())
	require.Empty(t, emptyRegistry.Metadatas())
	_, err := emptyRegistry.Get(ts.StringID("someString"))
	require.Error(t, err)
}

func TestRegistrySingleElement(t *testing.T) {
	var (
		opts      = NewOptions()
		id        = ts.StringID("someID")
		metadatas = []Metadata{
			NewMetadata(id, opts),
		}
		registry = NewRegistry(metadatas)
	)
	require.Equal(t, 1, len(registry.IDs()))
	require.Equal(t, id.String(), registry.IDs()[0].String())
	require.Equal(t, 1, len(registry.Metadatas()))
	md := registry.Metadatas()[0]
	require.Equal(t, id.String(), md.ID().String())
	require.Equal(t, opts, md.Options())
}

func TestRegistryMultipleElements(t *testing.T) {
	var (
		opts1     = NewOptions()
		opts2     = opts1.SetNeedsRepair(true)
		id1       = ts.StringID("someID1")
		id2       = ts.StringID("someID2")
		metadatas = []Metadata{
			NewMetadata(id1, opts1),
			NewMetadata(id2, opts2),
		}
		registry = NewRegistry(metadatas)
	)
	require.Equal(t, 2, len(registry.IDs()))
	ids := registry.IDs()
	require.True(t, ids[0].Equal(id1) || ids[1].Equal(id1))
	require.True(t, ids[0].Equal(id2) || ids[1].Equal(id2))

	require.Equal(t, 2, len(registry.Metadatas()))
	mds := registry.Metadatas()
	require.True(t, id1.Equal(mds[0].ID()) || id1.Equal(mds[1].ID()))
	require.True(t, id2.Equal(mds[0].ID()) || id2.Equal(mds[1].ID()))
}

func testRegistry() Registry {
	var (
		opts1     = NewOptions()
		opts2     = opts1.SetNeedsRepair(true)
		id1       = ts.StringID("someID1")
		id2       = ts.StringID("someID2")
		metadatas = []Metadata{
			NewMetadata(id1, opts1),
			NewMetadata(id2, opts2),
		}
	)
	return NewRegistry(metadatas)
}

func TestRegistryEqualsTrue(t *testing.T) {
	r1 := testRegistry()
	require.True(t, r1.Equal(r1))

	r2 := testRegistry()
	require.True(t, r1.Equal(r2))
	require.True(t, r2.Equal(r1))
}

func TestRegistryValidateSimple(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		opts      = NewMockOptions(ctrl)
		id        = ts.StringID("someID")
		metadatas = []Metadata{
			NewMetadata(id, opts),
		}
		reg = NewRegistry(metadatas)
	)

	opts.EXPECT().Validate().Return(nil)
	require.NoError(t, reg.Validate())

	opts.EXPECT().Validate().Return(fmt.Errorf("test error in options"))
	require.Error(t, reg.Validate())
}

func TestRegistryValidateDuplicateID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		opts      = NewMockOptions(ctrl)
		id        = ts.StringID("someID")
		metadatas = []Metadata{
			NewMetadata(id, opts),
			NewMetadata(id, opts),
		}
		reg = NewRegistry(metadatas)
	)

	opts.EXPECT().Validate().Return(nil).AnyTimes()
	require.Error(t, reg.Validate())
}