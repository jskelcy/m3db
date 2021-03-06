package clone

import (
	"fmt"
	"io"
	"time"

	"github.com/m3db/m3db/persist/encoding/msgpack"
	"github.com/m3db/m3db/persist/fs"
	"github.com/m3db/m3db/ts"
)

type cloner struct {
	opts Options
}

// New creates a new fileset cloner
func New(opts Options) FilesetCloner {
	return &cloner{
		opts: opts,
	}
}

func (c *cloner) Clone(src FilesetID, dest FilesetID, destBlocksize time.Duration) error {
	reader := fs.NewReader(src.PathPrefix, c.opts.BufferSize(), nil, msgpack.NewDecodingOptions())
	if err := reader.Open(ts.StringID(src.Namespace), src.Shard, src.Blockstart); err != nil {
		return fmt.Errorf("unable to read source fileset: %v", err)
	}

	writer := fs.NewWriter(dest.PathPrefix, c.opts.BufferSize(), c.opts.FileMode(), c.opts.DirMode())
	if err := writer.Open(ts.StringID(dest.Namespace), destBlocksize, dest.Shard, dest.Blockstart); err != nil {
		return fmt.Errorf("unable to open fileset writer: %v", err)
	}

	for {
		id, data, checksum, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("unexpected error while reading data: %v", err)
		}

		data.IncRef()
		if err := writer.Write(id, data, checksum); err != nil {
			return fmt.Errorf("unexpected error while writing data: %v", err)
		}
		data.DecRef()
		data.Finalize()
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("unable to finalize writer: %v", err)
	}

	if err := reader.Close(); err != nil {
		return fmt.Errorf("unable to finalize reader: %v", err)
	}

	return nil
}
