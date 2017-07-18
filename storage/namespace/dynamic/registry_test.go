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

package dynamic

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/m3db/m3cluster/client"
	"github.com/m3db/m3cluster/kv"
	nsproto "github.com/m3db/m3db/generated/proto/namespace"
	"github.com/m3db/m3db/storage/namespace"

	"github.com/fortytw2/leaktest"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/m3db/m3x/instrument"
	"github.com/stretchr/testify/require"
	"github.com/uber-go/tally"
)

func newTestOpts(ctrl *gomock.Controller, watch *testValueWatch) namespace.DynamicOptions {
	// update report interval for tests
	defaultReportInterval = 10 * time.Millisecond

	ts := tally.NewTestScope("", nil)
	mockKVStore := kv.NewMockStore(ctrl)
	mockKVStore.EXPECT().Watch(defaultNsRegistryKey).Return(watch, nil)

	mockCSClient := client.NewMockClient(ctrl)
	mockCSClient.EXPECT().KV().Return(mockKVStore, nil)

	opts := NewOptions().
		SetInstrumentOptions(
			instrument.NewOptions().SetMetricsScope(ts)).
		SetInitTimeout(100 * time.Millisecond).
		SetConfigServiceClient(mockCSClient)

	return opts
}

func TestInitializerTimeout(t *testing.T) {
	defer leaktest.CheckTimeout(t, time.Second)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	values := singleTestValue()
	w := newWatch(values)
	opts := newTestOpts(ctrl, w)
	init := NewInitializer(opts)
	_, err := init.Init()
	require.Error(t, err)
	require.Equal(t, errInitTimeOut, err)
}

func TestInitializerNoTimeout(t *testing.T) {
	defer leaktest.CheckTimeout(t, time.Second)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	value := singleTestValue()
	expectedNsValue := value.Namespaces["testns1"]
	w := newWatch(value)
	defer w.Close()
	go w.start()

	opts := newTestOpts(ctrl, w)
	init := NewInitializer(opts)
	reg, err := init.Init()
	require.NoError(t, err)

	rw, err := reg.Watch()
	require.NoError(t, err)
	rMap := rw.Get()
	mds := rMap.Metadatas()
	require.Len(t, mds, 1)
	md := mds[0]
	require.Equal(t, "testns1", md.ID().String())
	require.Equal(t, expectedNsValue.NeedsBootstrap, md.Options().NeedsBootstrap())
	require.Equal(t, expectedNsValue.NeedsFilesetCleanup, md.Options().NeedsFilesetCleanup())
	require.Equal(t, expectedNsValue.NeedsFlush, md.Options().NeedsFlush())
	require.Equal(t, expectedNsValue.NeedsRepair, md.Options().NeedsRepair())
	require.Equal(t, expectedNsValue.WritesToCommitLog, md.Options().WritesToCommitLog())

	ropts := expectedNsValue.RetentionOptions
	observedRopts := md.Options().RetentionOptions()
	require.Equal(t, ropts.BlockDataExpiry, observedRopts.BlockDataExpiry())
	require.Equal(t, ropts.BlockDataExpiryAfterNotAccessPeriodNanos,
		observedRopts.BlockDataExpiryAfterNotAccessedPeriod().Nanoseconds())
	require.Equal(t, ropts.BlockSizeNanos, observedRopts.BlockSize().Nanoseconds())
	require.Equal(t, ropts.BufferFutureNanos, observedRopts.BufferFuture().Nanoseconds())
	require.Equal(t, ropts.BufferPastNanos, observedRopts.BufferPast().Nanoseconds())

	require.NoError(t, rw.Close())
	require.NoError(t, reg.Close())
}

func TestInitializerUpdateWithBadProto(t *testing.T) {
	defer leaktest.CheckTimeout(t, time.Second)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	w := newWatch(singleTestValue())
	defer w.Close()

	opts := newTestOpts(ctrl, w)
	init := NewInitializer(opts)

	scope := opts.InstrumentOptions().MetricsScope().(tally.TestScope)
	numInvalid := func() int64 {
		count, ok := scope.Snapshot().Counters()["namespace_registry.invalid_update+"]
		if !ok {
			return 0
		}
		return count.Value()
	}

	go w.start()
	reg, err := init.Init()
	require.NoError(t, err)

	rmap, err := reg.Watch()
	require.NoError(t, err)
	require.Len(t, rmap.Get().Metadatas(), 1)
	require.Equal(t, int64(0), numInvalid())

	// update with bad proto
	w.valueCh <- testValue{
		version: 2,
		Registry: nsproto.Registry{
			Namespaces: map[string]*nsproto.NamespaceOptions{
				"testns1": nil,
				"testns2": nil,
			},
		},
	}

	time.Sleep(20 * time.Millisecond)
	require.Equal(t, int64(1), numInvalid())

	require.Len(t, rmap.Get().Metadatas(), 1)
	require.NoError(t, reg.Close())
}

func TestInitializerUpdateWithOlderVersion(t *testing.T) {
	defer leaktest.CheckTimeout(t, time.Second)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	initValue := singleTestValue()
	w := newWatch(initValue)
	defer w.Close()

	opts := newTestOpts(ctrl, w)
	init := NewInitializer(opts)

	scope := opts.InstrumentOptions().MetricsScope().(tally.TestScope)
	numInvalid := func() int64 {
		count, ok := scope.Snapshot().Counters()["namespace_registry.invalid_update+"]
		if !ok {
			return 0
		}
		return count.Value()
	}

	go w.start()
	reg, err := init.Init()
	require.NoError(t, err)

	rmap, err := reg.Watch()
	require.NoError(t, err)
	require.Len(t, rmap.Get().Metadatas(), 1)
	require.Equal(t, int64(0), numInvalid())

	// update with bad version
	w.valueCh <- testValue{
		version:  1,
		Registry: initValue.Registry,
	}

	time.Sleep(20 * time.Millisecond)
	require.Equal(t, int64(1), numInvalid())

	require.Len(t, rmap.Get().Metadatas(), 1)
	require.NoError(t, reg.Close())
}

func TestInitializerUpdateWithNewerVersion(t *testing.T) {
	defer leaktest.CheckTimeout(t, time.Second)()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	initValue := singleTestValue()
	w := newWatch(initValue)
	defer w.Close()

	opts := newTestOpts(ctrl, w)
	init := NewInitializer(opts)

	scope := opts.InstrumentOptions().MetricsScope().(tally.TestScope)
	numInvalid := func() int64 {
		count, ok := scope.Snapshot().Counters()["namespace_registry.invalid_update+"]
		if !ok {
			return 0
		}
		return count.Value()
	}
	currentVersion := func() float64 {
		g, ok := scope.Snapshot().Gauges()["namespace_registry.current_version+"]
		if !ok {
			return 0.0
		}
		return g.Value()
	}

	go w.start()
	reg, err := init.Init()
	require.NoError(t, err)

	rmap, err := reg.Watch()
	require.NoError(t, err)
	require.Len(t, rmap.Get().Metadatas(), 1)
	require.Equal(t, int64(0), numInvalid())
	require.Equal(t, 0., currentVersion())

	// update with new version
	w.valueCh <- testValue{
		version:  2,
		Registry: initValue.Registry,
	}

	time.Sleep(20 * time.Millisecond)
	require.Equal(t, int64(0), numInvalid())
	require.Equal(t, 2., currentVersion())

	require.Len(t, rmap.Get().Metadatas(), 1)
	require.NoError(t, reg.Close())
}

func singleTestValue() testValue {
	return testValue{
		version: 1,
		Registry: nsproto.Registry{
			Namespaces: map[string]*nsproto.NamespaceOptions{
				"testns1": &nsproto.NamespaceOptions{
					NeedsBootstrap:      true,
					NeedsFilesetCleanup: true,
					NeedsFlush:          true,
					NeedsRepair:         true,
					WritesToCommitLog:   true,
					RetentionOptions: &nsproto.RetentionOptions{
						BlockDataExpiry:                          true,
						BlockDataExpiryAfterNotAccessPeriodNanos: int64(time.Minute),
						BlockSizeNanos:                           int64(time.Hour * 2),
						RetentionPeriodNanos:                     int64(time.Hour * 48),
						BufferFutureNanos:                        int64(time.Minute * 10),
						BufferPastNanos:                          int64(time.Minute * 15),
					},
				},
			},
		},
	}
}

type testValue struct {
	nsproto.Registry
	version int
}

func (v *testValue) Unmarshal(msg proto.Message) error {
	reg, ok := msg.(*nsproto.Registry)
	if !ok {
		return fmt.Errorf("incorrect type provided: %T", msg)
	}
	reg.Namespaces = v.Namespaces
	return nil
}

func (v *testValue) Version() int {
	return v.version
}

func (v *testValue) IsNewer(other kv.Value) bool {
	return v.Version() > other.Version()
}

type testValueWatch struct {
	notifyCh chan struct{}
	valueCh  chan testValue

	initValue testValue

	sync.RWMutex
	current *testValue
	closed  bool
}

func newWatch(initValue testValue) *testValueWatch {
	return &testValueWatch{
		initValue: initValue,
		notifyCh:  make(chan struct{}, 10),
		valueCh:   make(chan testValue, 10),
	}
}

func (w *testValueWatch) updateValue(v testValue) {
	w.Lock()
	w.current = &v
	w.notifyCh <- struct{}{}
	w.Unlock()
}

func (w *testValueWatch) start() {
	w.updateValue(w.initValue)
	for val := range w.valueCh {
		w.updateValue(val)
	}
}

func (w *testValueWatch) C() <-chan struct{} {
	return w.notifyCh
}

func (w *testValueWatch) Get() kv.Value {
	w.RLock()
	defer w.RUnlock()
	return w.current
}

func (w *testValueWatch) Close() {
	w.Lock()
	if w.closed {
		w.Unlock()
		return
	}
	w.closed = true
	w.Unlock()

	close(w.valueCh)
	close(w.notifyCh)
}