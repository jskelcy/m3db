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

package storage

import (
	"errors"
	"sync"
	"time"

	"github.com/m3db/m3db/storage/namespace"
	"github.com/m3db/m3x/instrument"
	xlog "github.com/m3db/m3x/log"

	"github.com/uber-go/tally"
)

var (
	errAlreadyWatching = errors.New("database is already watching namespace updates")
	errNotWatching     = errors.New("databse is not watching for namespace updates")
)

type dbNamespaceWatch struct {
	sync.Mutex
	watching bool
	watch    namespace.Watch
	doneCh   chan struct{}
	closedCh chan struct{}

	db  database
	log xlog.Logger

	metrics dbNamespaceWatchMetrics
}

type dbNamespaceWatchMetrics struct {
	activeNamespaces tally.Gauge
	updates          tally.Counter
}

func newWatchMetrics(
	scope tally.Scope,
) dbNamespaceWatchMetrics {
	nsScope := scope.SubScope("database.nswatch")
	return dbNamespaceWatchMetrics{
		activeNamespaces: nsScope.Gauge("active"),
		updates:          nsScope.Counter("updates"),
	}
}

func newDatabaseNamespaceWatch(
	db database,
	w namespace.Watch,
	iopts instrument.Options,
) databaseNamespaceWatch {
	scope := iopts.MetricsScope()
	return &dbNamespaceWatch{
		watch:   w,
		db:      db,
		log:     iopts.Logger(),
		metrics: newWatchMetrics(scope),
	}
}

// TODO(prateek): write tests for databaseNamespaceWatch

func (w *dbNamespaceWatch) Start() error {
	w.Lock()
	defer w.Unlock()

	if w.watching {
		return errAlreadyWatching
	}

	w.doneCh = make(chan struct{}, 1)
	w.closedCh = make(chan struct{}, 1)

	go w.startWatch()

	return nil
}

func (w *dbNamespaceWatch) startWatch() {
	reportClosingCh := make(chan struct{}, 1)
	reportClosedCh := make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				w.reportMetrics()
			case <-reportClosingCh:
				ticker.Stop()
				close(reportClosedCh)
				return
			}
		}
	}()

	for {
		select {
		case <-w.doneCh:
			// Issue closing signal to report channel
			close(reportClosingCh)
			// Wait for report channel to close
			<-reportClosedCh
			// Signal all closed
			close(w.closedCh)
			return
		case <-w.watch.C():
			w.metrics.updates.Inc(1)
			w.db.updateOwnedNamespaces(w.watch.Get())
		}
	}
}

func (w *dbNamespaceWatch) reportMetrics() {
	mds := w.watch.Get().Metadatas()
	w.metrics.activeNamespaces.Update(float64(len(mds)))
}

func (w *dbNamespaceWatch) Stop() error {
	w.Lock()
	defer w.Unlock()

	if !w.watching {
		return errNotWatching
	}

	w.watching = false

	close(w.doneCh)
	<-w.closedCh

	return nil
}

func (w *dbNamespaceWatch) Close() error {
	w.Lock()
	watching := w.watching
	w.Unlock()
	if watching {
		return w.Stop()
	}
	return nil
}