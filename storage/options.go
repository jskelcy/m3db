// Copyright (c) 2016 Uber Technologies, Inc.
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

package storage

import (
	"io"
	"time"

	"github.com/m3db/m3db/clock"
	"github.com/m3db/m3db/context"
	"github.com/m3db/m3db/encoding"
	"github.com/m3db/m3db/encoding/m3tsz"
	"github.com/m3db/m3db/persist"
	"github.com/m3db/m3db/persist/fs"
	"github.com/m3db/m3db/persist/fs/commitlog"
	"github.com/m3db/m3db/retention"
	"github.com/m3db/m3db/storage/block"
	"github.com/m3db/m3db/storage/bootstrap"
	"github.com/m3db/m3db/storage/repair"
	"github.com/m3db/m3db/storage/series"
	"github.com/m3db/m3db/ts"
	xio "github.com/m3db/m3db/x/io"
	"github.com/m3db/m3x/instrument"
	"github.com/m3db/m3x/pool"
)

const (
	// defaultCommitLogStrategy is the default commit log strategy
	defaultCommitLogStrategy = commitlog.StrategyWriteBehind

	// defaultMaxFlushRetries is the default number of retries when flush fails
	defaultMaxFlushRetries = 3

	// defaultShardCloseDeadline is the default soft deadline to give when closing shards,
	// it is used to gate the impact of closing the shard on the performnace of the process.
	// It by default is 2 minutes which aligns with the Go GC timer so it should at least be
	// split across two GC cycles somewhat.
	defaultShardCloseDeadline = 2 * time.Minute

	// defaultBytesPoolBucketCapacity is the default bytes buffer capacity for the default bytes pool bucket
	defaultBytesPoolBucketCapacity = 256

	// defaultBytesPoolBucketCount is the default count of elements for the default bytes pool bucket
	defaultBytesPoolBucketCount = 4096
)

var (
	// defaultNewBootstrapFn is the default function for creating a new bootstrap
	defaultNewBootstrapFn = func() bootstrap.Bootstrap {
		return bootstrap.NewNoOpBootstrapProcess()
	}

	// defaultNewPersistManagerFn is the default function for creating a new persist manager
	defaultNewPersistManagerFn = func() persist.Manager {
		return fs.NewPersistManager(fs.NewOptions())
	}

	timeZero time.Time
)

// NewSeriesOptionsFromOptions creates a new set of database series options from options
func NewSeriesOptionsFromOptions(opts Options) series.Options {
	return series.NewOptions().
		SetClockOptions(opts.ClockOptions()).
		SetInstrumentOptions(opts.InstrumentOptions()).
		SetRetentionOptions(opts.RetentionOptions()).
		SetDatabaseBlockOptions(opts.DatabaseBlockOptions()).
		SetContextPool(opts.ContextPool()).
		SetEncoderPool(opts.EncoderPool()).
		SetMultiReaderIteratorPool(opts.MultiReaderIteratorPool()).
		SetIdentifierPool(opts.IdentifierPool())
}

type options struct {
	clockOpts                      clock.Options
	instrumentOpts                 instrument.Options
	retentionOpts                  retention.Options
	blockOpts                      block.Options
	commitLogOpts                  commitlog.Options
	repairOpts                     repair.Options
	fileOpOpts                     FileOpOptions
	newEncoderFn                   encoding.NewEncoderFn
	newDecoderFn                   encoding.NewDecoderFn
	newBootstrapFn                 NewBootstrapFn
	newPersistManagerFn            NewPersistManagerFn
	maxFlushRetries                int
	shardCloseDeadline             time.Duration
	contextPool                    context.Pool
	seriesPool                     series.DatabaseSeriesPool
	bytesPool                      pool.CheckedBytesPool
	encoderPool                    encoding.EncoderPool
	segmentReaderPool              xio.SegmentReaderPool
	readerIteratorPool             encoding.ReaderIteratorPool
	multiReaderIteratorPool        encoding.MultiReaderIteratorPool
	identifierPool                 ts.IdentifierPool
	fetchBlockMetadataResultsPool  block.FetchBlockMetadataResultsPool
	fetchBlocksMetadataResultsPool block.FetchBlocksMetadataResultsPool
}

// NewOptions creates a new set of storage options with defaults
// TODO(r): add an "IsValid()" method and ensure buffer future and buffer past are
// less than blocksize and check when opening database
func NewOptions() Options {
	bytesPool := pool.NewCheckedBytesPool(nil, nil, func(s []pool.Bucket) pool.BytesPool {
		return pool.NewBytesPool(s, nil)
	})
	o := &options{
		clockOpts:                      clock.NewOptions(),
		instrumentOpts:                 instrument.NewOptions(),
		retentionOpts:                  retention.NewOptions(),
		blockOpts:                      block.NewOptions(),
		commitLogOpts:                  commitlog.NewOptions(),
		repairOpts:                     repair.NewOptions(),
		fileOpOpts:                     NewFileOpOptions(),
		newBootstrapFn:                 defaultNewBootstrapFn,
		newPersistManagerFn:            defaultNewPersistManagerFn,
		maxFlushRetries:                defaultMaxFlushRetries,
		shardCloseDeadline:             defaultShardCloseDeadline,
		contextPool:                    context.NewPool(nil, nil),
		seriesPool:                     series.NewDatabaseSeriesPool(series.NewOptions(), nil),
		bytesPool:                      bytesPool,
		encoderPool:                    encoding.NewEncoderPool(nil),
		segmentReaderPool:              xio.NewSegmentReaderPool(nil),
		readerIteratorPool:             encoding.NewReaderIteratorPool(nil),
		multiReaderIteratorPool:        encoding.NewMultiReaderIteratorPool(nil),
		identifierPool:                 ts.NewIdentifierPool(bytesPool, nil),
		fetchBlockMetadataResultsPool:  block.NewFetchBlockMetadataResultsPool(nil, 0),
		fetchBlocksMetadataResultsPool: block.NewFetchBlocksMetadataResultsPool(nil, 0),
	}
	return o.SetEncodingM3TSZPooled()
}

func (o *options) SetClockOptions(value clock.Options) Options {
	opts := *o
	opts.clockOpts = value
	opts.commitLogOpts = opts.commitLogOpts.SetClockOptions(value)
	opts.seriesPool = series.NewDatabaseSeriesPool(NewSeriesOptionsFromOptions(&opts), nil)
	return &opts
}

func (o *options) ClockOptions() clock.Options {
	return o.clockOpts
}

func (o *options) SetInstrumentOptions(value instrument.Options) Options {
	opts := *o
	opts.instrumentOpts = value
	opts.commitLogOpts = opts.commitLogOpts.SetInstrumentOptions(value)
	opts.seriesPool = series.NewDatabaseSeriesPool(NewSeriesOptionsFromOptions(&opts), nil)
	return &opts
}

func (o *options) InstrumentOptions() instrument.Options {
	return o.instrumentOpts
}

func (o *options) SetRetentionOptions(value retention.Options) Options {
	opts := *o
	opts.retentionOpts = value
	opts.commitLogOpts = opts.commitLogOpts.SetRetentionOptions(value)
	opts.seriesPool = series.NewDatabaseSeriesPool(NewSeriesOptionsFromOptions(&opts), nil)
	return &opts
}

func (o *options) RetentionOptions() retention.Options {
	return o.retentionOpts
}

func (o *options) SetDatabaseBlockOptions(value block.Options) Options {
	opts := *o
	opts.blockOpts = value
	opts.seriesPool = series.NewDatabaseSeriesPool(NewSeriesOptionsFromOptions(&opts), nil)
	return &opts
}

func (o *options) DatabaseBlockOptions() block.Options {
	return o.blockOpts
}

func (o *options) SetCommitLogOptions(value commitlog.Options) Options {
	opts := *o
	opts.commitLogOpts = value
	return &opts
}

func (o *options) CommitLogOptions() commitlog.Options {
	return o.commitLogOpts
}

func (o *options) SetRepairOptions(value repair.Options) Options {
	opts := *o
	opts.repairOpts = value
	return &opts
}

func (o *options) RepairOptions() repair.Options {
	return o.repairOpts
}

func (o *options) SetFileOpOptions(value FileOpOptions) Options {
	opts := *o
	opts.fileOpOpts = value
	return &opts
}

func (o *options) FileOpOptions() FileOpOptions {
	return o.fileOpOpts
}

func (o *options) SetEncodingM3TSZPooled() Options {
	opts := *o

	buckets := []pool.Bucket{{
		Capacity: defaultBytesPoolBucketCapacity,
		Count:    defaultBytesPoolBucketCount,
	}}
	newBackingBytesPool := func(s []pool.Bucket) pool.BytesPool {
		return pool.NewBytesPool(s, nil)
	}
	bytesPool := pool.NewCheckedBytesPool(buckets, nil, newBackingBytesPool)
	bytesPool.Init()
	opts.bytesPool = bytesPool

	// initialize context pool
	opts.contextPool = context.NewPool(nil, nil)

	encoderPool := encoding.NewEncoderPool(nil)
	readerIteratorPool := encoding.NewReaderIteratorPool(nil)

	// initialize segment reader pool
	segmentReaderPool := xio.NewSegmentReaderPool(nil)
	segmentReaderPool.Init()
	opts.segmentReaderPool = segmentReaderPool

	encodingOpts := encoding.NewOptions().
		SetBytesPool(bytesPool).
		SetEncoderPool(encoderPool).
		SetReaderIteratorPool(readerIteratorPool).
		SetSegmentReaderPool(segmentReaderPool)

	// initialize encoder pool
	encoderPool.Init(func() encoding.Encoder {
		return m3tsz.NewEncoder(timeZero, nil, m3tsz.DefaultIntOptimizationEnabled, encodingOpts)
	})
	opts.encoderPool = encoderPool

	// initialize single reader iterator pool
	readerIteratorPool.Init(func(r io.Reader) encoding.ReaderIterator {
		return m3tsz.NewReaderIterator(r, m3tsz.DefaultIntOptimizationEnabled, encodingOpts)
	})
	opts.readerIteratorPool = readerIteratorPool

	// initialize multi reader iterator pool
	multiReaderIteratorPool := encoding.NewMultiReaderIteratorPool(nil)
	multiReaderIteratorPool.Init(func(r io.Reader) encoding.ReaderIterator {
		return m3tsz.NewReaderIterator(r, m3tsz.DefaultIntOptimizationEnabled, encodingOpts)
	})
	opts.multiReaderIteratorPool = multiReaderIteratorPool

	opts.blockOpts = opts.blockOpts.
		SetEncoderPool(encoderPool).
		SetReaderIteratorPool(readerIteratorPool).
		SetMultiReaderIteratorPool(multiReaderIteratorPool)

	opts.seriesPool = series.NewDatabaseSeriesPool(NewSeriesOptionsFromOptions(&opts), nil)

	return &opts
}

func (o *options) SetNewEncoderFn(value encoding.NewEncoderFn) Options {
	opts := *o
	opts.newEncoderFn = value
	return &opts
}

func (o *options) NewEncoderFn() encoding.NewEncoderFn {
	return o.newEncoderFn
}

func (o *options) SetNewDecoderFn(value encoding.NewDecoderFn) Options {
	opts := *o
	opts.newDecoderFn = value
	return &opts
}

func (o *options) NewDecoderFn() encoding.NewDecoderFn {
	return o.newDecoderFn
}

func (o *options) SetNewBootstrapFn(value NewBootstrapFn) Options {
	opts := *o
	opts.newBootstrapFn = value
	return &opts
}

func (o *options) NewBootstrapFn() NewBootstrapFn {
	return o.newBootstrapFn
}

func (o *options) SetNewPersistManagerFn(value NewPersistManagerFn) Options {
	opts := *o
	opts.newPersistManagerFn = value
	return &opts
}

func (o *options) NewPersistManagerFn() NewPersistManagerFn {
	return o.newPersistManagerFn
}

func (o *options) SetMaxFlushRetries(value int) Options {
	opts := *o
	opts.maxFlushRetries = value
	return &opts
}

func (o *options) MaxFlushRetries() int {
	return o.maxFlushRetries
}

func (o *options) SetShardCloseDeadline(value time.Duration) Options {
	opts := *o
	opts.shardCloseDeadline = value
	return &opts
}

func (o *options) ShardCloseDeadline() time.Duration {
	return o.shardCloseDeadline
}

func (o *options) SetContextPool(value context.Pool) Options {
	opts := *o
	opts.contextPool = value
	return &opts
}

func (o *options) ContextPool() context.Pool {
	return o.contextPool
}

func (o *options) SetDatabaseSeriesPool(value series.DatabaseSeriesPool) Options {
	opts := *o
	opts.seriesPool = value
	return &opts
}

func (o *options) DatabaseSeriesPool() series.DatabaseSeriesPool {
	return o.seriesPool
}

func (o *options) SetBytesPool(value pool.CheckedBytesPool) Options {
	opts := *o
	opts.bytesPool = value
	return &opts
}

func (o *options) BytesPool() pool.CheckedBytesPool {
	return o.bytesPool
}

func (o *options) SetEncoderPool(value encoding.EncoderPool) Options {
	opts := *o
	opts.encoderPool = value
	return &opts
}

func (o *options) EncoderPool() encoding.EncoderPool {
	return o.encoderPool
}

func (o *options) SetSegmentReaderPool(value xio.SegmentReaderPool) Options {
	opts := *o
	opts.segmentReaderPool = value
	return &opts
}

func (o *options) SegmentReaderPool() xio.SegmentReaderPool {
	return o.segmentReaderPool
}

func (o *options) SetReaderIteratorPool(value encoding.ReaderIteratorPool) Options {
	opts := *o
	opts.readerIteratorPool = value
	return &opts
}

func (o *options) ReaderIteratorPool() encoding.ReaderIteratorPool {
	return o.readerIteratorPool
}

func (o *options) SetMultiReaderIteratorPool(value encoding.MultiReaderIteratorPool) Options {
	opts := *o
	opts.multiReaderIteratorPool = value
	return &opts
}

func (o *options) MultiReaderIteratorPool() encoding.MultiReaderIteratorPool {
	return o.multiReaderIteratorPool
}

func (o *options) SetIdentifierPool(value ts.IdentifierPool) Options {
	opts := *o
	opts.identifierPool = value
	return &opts
}

func (o *options) IdentifierPool() ts.IdentifierPool {
	return o.identifierPool
}

func (o *options) SetFetchBlockMetadataResultsPool(value block.FetchBlockMetadataResultsPool) Options {
	opts := *o
	opts.fetchBlockMetadataResultsPool = value
	return &opts
}

func (o *options) FetchBlockMetadataResultsPool() block.FetchBlockMetadataResultsPool {
	return o.fetchBlockMetadataResultsPool
}

func (o *options) SetFetchBlocksMetadataResultsPool(value block.FetchBlocksMetadataResultsPool) Options {
	opts := *o
	opts.fetchBlocksMetadataResultsPool = value
	return &opts
}

func (o *options) FetchBlocksMetadataResultsPool() block.FetchBlocksMetadataResultsPool {
	return o.fetchBlocksMetadataResultsPool
}
