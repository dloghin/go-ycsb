// Copyright 2018 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package pebble

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/bloom"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/magiconair/properties"
	"github.com/pingcap/go-ycsb/pkg/prop"
	"github.com/pingcap/go-ycsb/pkg/util"
	"github.com/pingcap/go-ycsb/pkg/ycsb"
)

// properties
const (
	pebbleDir                         = "pebble.dir"
	pebbleBytesPerSync                = "pebble.bytes_per_sync"
	pebbleCacheSize                   = "pebble.cache_size"
	pebbleDisableWAL                  = "pebble.disable_wal"
	pebbleErrorIfExists               = "pebble.error_if_exists"
	pebbleErrorIfNotExists            = "pebble.error_if_not_exists"
	pebbleErrorIfNotPristine          = "pebble.error_if_not_pristine"
	pebbleFlushDelayDeleteRange       = "pebble.flush_delay_delete_range"
	pebbleFlushDelayRangeKey          = "pebble.flush_delay_range_key"
	pebbleFlushSplitBytes             = "pebble.flush_split_bytes"
	pebbleL0CompactionFileThreshold   = "pebble.l0_compaction_file_threshold"
	pebbleL0CompactionThreshold       = "pebble.l0_compaction_threshold"
	pebbleL0StopWritesThreshold       = "pebble.l0_stop_writes_threshold"
	pebbleLBaseMaxBytes               = "pebble.lbase_max_bytes"
	pebbleMaxManifestFileSize         = "pebble.max_manifest_file_size"
	pebbleMaxOpenFiles                = "pebble.max_open_files"
	pebbleMemTableSize                = "pebble.memtable_size"
	pebbleMemTableStopWritesThreshold = "pebble.memtable_stop_writes_threshold"
	pebbleMaxConcurrentCompactions    = "pebble.max_concurrent_compactions"
	pebbleDisableAutomaticCompactions = "pebble.disable_automatic_compactions"
	pebbleNoSyncOnClose               = "pebble.no_sync_on_close"
	pebbleNumPrevManifest             = "pebble.num_prev_manifest"
	pebbleReadOnly                    = "pebble.read_only"
	pebbleWALBytesPerSync             = "pebble.wal_bytes_per_sync"
	pebbleWALDir                      = "pebble.wal_dir"
	pebbleWALMinSyncInterval          = "pebble.wal_min_sync_interval"
	pebbleTargetByteDeletionRate      = "pebble.target_byte_deletion_rate"

	// Level options
	pebbleLevelBlockRestartInterval = "pebble.level.block_restart_interval"
	pebbleLevelBlockSize            = "pebble.level.block_size"
	pebbleLevelBlockSizeThreshold   = "pebble.level.block_size_threshold"
	pebbleLevelCompression          = "pebble.level.compression"
	pebbleLevelFilterPolicy         = "pebble.level.filter_policy"
	pebbleLevelFilterType           = "pebble.level.filter_type"
	pebbleLevelIndexBlockSize       = "pebble.level.index_block_size"
	pebbleLevelTargetFileSize       = "pebble.level.target_file_size"
)

type pebbleCreator struct{}

type pebbleDB struct {
	p *properties.Properties

	db *pebble.DB

	r       *util.RowCodec
	bufPool *util.BufPool

	readOpts  *pebble.IterOptions
	writeOpts *pebble.WriteOptions
}

func (c pebbleCreator) Create(p *properties.Properties) (ycsb.DB, error) {
	dir := p.GetString(pebbleDir, "/tmp/pebble")

	if p.GetBool(prop.DropData, prop.DropDataDefault) {
		os.RemoveAll(dir)
	}

	opts := getOptions(p)

	db, err := pebble.Open(dir, opts)
	if err != nil {
		return nil, err
	}

	return &pebbleDB{
		p:         p,
		db:        db,
		r:         util.NewRowCodec(p),
		bufPool:   util.NewBufPool(),
		readOpts:  getReadOptions(p),
		writeOpts: getWriteOptions(p),
	}, nil
}

func getOptions(p *properties.Properties) *pebble.Options {
	opts := &pebble.Options{}

	// Basic options
	opts.BytesPerSync = p.GetInt(pebbleBytesPerSync, 512*1024) // 512KB default

	// Cache settings
	cacheSize := p.GetInt(pebbleCacheSize, 8*1024*1024) // 8MB default
	if cacheSize > 0 {
		opts.Cache = pebble.NewCache(int64(cacheSize))
	}

	// WAL settings
	opts.DisableWAL = p.GetBool(pebbleDisableWAL, false)
	opts.WALBytesPerSync = p.GetInt(pebbleWALBytesPerSync, 0)
	opts.WALDir = p.GetString(pebbleWALDir, "")

	// WAL min sync interval
	if walMinSync := p.GetInt(pebbleWALMinSyncInterval, 0); walMinSync > 0 {
		opts.WALMinSyncInterval = func() time.Duration {
			return time.Duration(walMinSync) * time.Microsecond
		}
	}

	// Error conditions
	opts.ErrorIfExists = p.GetBool(pebbleErrorIfExists, false)
	opts.ErrorIfNotExists = p.GetBool(pebbleErrorIfNotExists, false)
	opts.ErrorIfNotPristine = p.GetBool(pebbleErrorIfNotPristine, false)

	// Flush settings
	if flushDelayDeleteRange := p.GetInt(pebbleFlushDelayDeleteRange, 0); flushDelayDeleteRange > 0 {
		opts.FlushDelayDeleteRange = time.Duration(flushDelayDeleteRange) * time.Millisecond
	}
	if flushDelayRangeKey := p.GetInt(pebbleFlushDelayRangeKey, 0); flushDelayRangeKey > 0 {
		opts.FlushDelayRangeKey = time.Duration(flushDelayRangeKey) * time.Millisecond
	}
	opts.FlushSplitBytes = p.GetInt64(pebbleFlushSplitBytes, 0)

	// L0 settings
	opts.L0CompactionFileThreshold = p.GetInt(pebbleL0CompactionFileThreshold, 4)
	opts.L0CompactionThreshold = p.GetInt(pebbleL0CompactionThreshold, 4)
	opts.L0StopWritesThreshold = p.GetInt(pebbleL0StopWritesThreshold, 12)

	// Base level settings
	opts.LBaseMaxBytes = p.GetInt64(pebbleLBaseMaxBytes, 64*1024*1024) // 64MB default

	// File settings
	opts.MaxManifestFileSize = p.GetInt64(pebbleMaxManifestFileSize, 128*1024*1024) // 128MB default
	opts.MaxOpenFiles = p.GetInt(pebbleMaxOpenFiles, 1000)

	// MemTable settings
	opts.MemTableSize = p.GetUint64(pebbleMemTableSize, 4*1024*1024) // 4MB default
	opts.MemTableStopWritesThreshold = p.GetInt(pebbleMemTableStopWritesThreshold, 2)

	// Compaction settings
	maxConcurrentCompactions := p.GetInt(pebbleMaxConcurrentCompactions, 1)
	opts.MaxConcurrentCompactions = func() int {
		return maxConcurrentCompactions
	}
	opts.DisableAutomaticCompactions = p.GetBool(pebbleDisableAutomaticCompactions, false)

	// Sync settings
	opts.NoSyncOnClose = p.GetBool(pebbleNoSyncOnClose, false)
	opts.NumPrevManifest = p.GetInt(pebbleNumPrevManifest, 1)

	// Read-only mode
	opts.ReadOnly = p.GetBool(pebbleReadOnly, false)

	// Deletion rate limiting
	opts.TargetByteDeletionRate = p.GetInt(pebbleTargetByteDeletionRate, 0)

	// Set default levels with customizable options
	opts.Levels = getLevelOptions(p)

	// Use default filesystem
	opts.FS = vfs.Default

	return opts
}

func getLevelOptions(p *properties.Properties) []pebble.LevelOptions {
	// Create default level options
	levels := make([]pebble.LevelOptions, 7)
	for i := range levels {
		levels[i] = pebble.LevelOptions{
			BlockRestartInterval: p.GetInt(pebbleLevelBlockRestartInterval, 16),
			BlockSize:            p.GetInt(pebbleLevelBlockSize, 4096),
			BlockSizeThreshold:   p.GetInt(pebbleLevelBlockSizeThreshold, 90),
			IndexBlockSize:       p.GetInt(pebbleLevelIndexBlockSize, 4096),
			TargetFileSize:       p.GetInt64(pebbleLevelTargetFileSize, 2*1024*1024), // 2MB default
		}

		// Compression settings
		compression := p.GetString(pebbleLevelCompression, "snappy")
		switch compression {
		case "none":
			levels[i].Compression = pebble.NoCompression
		case "snappy":
			levels[i].Compression = pebble.SnappyCompression
		case "zstd":
			levels[i].Compression = pebble.ZstdCompression
		default:
			levels[i].Compression = pebble.SnappyCompression
		}

		// Filter policy
		if filterPolicy := p.GetString(pebbleLevelFilterPolicy, ""); filterPolicy == "bloom" {
			filterType := p.GetString(pebbleLevelFilterType, "table")
			if filterType == "block" {
				levels[i].FilterPolicy = bloom.FilterPolicy(10) // 10 bits per key
			} else {
				levels[i].FilterPolicy = bloom.FilterPolicy(10) // table filter is default
			}
		}
	}

	return levels
}

func getReadOptions(p *properties.Properties) *pebble.IterOptions {
	return &pebble.IterOptions{}
}

func getWriteOptions(p *properties.Properties) *pebble.WriteOptions {
	return &pebble.WriteOptions{
		Sync: true, // Default to sync writes for durability
	}
}

func (db *pebbleDB) Close() error {
	return db.db.Close()
}

func (db *pebbleDB) InitThread(ctx context.Context, _ int, _ int) context.Context {
	return ctx
}

func (db *pebbleDB) CleanupThread(_ context.Context) {
}

func (db *pebbleDB) getRowKey(table string, key string) []byte {
	return util.Slice(fmt.Sprintf("%s:%s", table, key))
}

func (db *pebbleDB) Read(ctx context.Context, table string, key string, fields []string) (map[string][]byte, error) {
	value, closer, err := db.db.Get(db.getRowKey(table, key))
	if err != nil {
		if err == pebble.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	defer func() {
		if closer != nil {
			closer.Close()
		}
	}()

	// Make a copy of the value since it's only valid until closer.Close()
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	return db.r.Decode(valueCopy, fields)
}

func (db *pebbleDB) Scan(ctx context.Context, table string, startKey string, count int, fields []string) ([]map[string][]byte, error) {
	res := make([]map[string][]byte, 0, count)

	iter, err := db.db.NewIter(db.readOpts)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	rowStartKey := db.getRowKey(table, startKey)
	tablePrefix := []byte(table + ":")

	// Start from the specified key
	if !iter.SeekGE(rowStartKey) {
		return res, nil
	}

	i := 0
	for iter.Valid() && i < count {
		key := iter.Key()

		// Check if the key belongs to the same table
		if !hasPrefix(key, tablePrefix) {
			break
		}

		value := iter.Value()

		// Make a copy of the value since the iterator might reuse the buffer
		valueCopy := make([]byte, len(value))
		copy(valueCopy, value)

		m, err := db.r.Decode(valueCopy, fields)
		if err != nil {
			return nil, err
		}

		res = append(res, m)
		i++

		if !iter.Next() {
			break
		}
	}

	if err := iter.Error(); err != nil {
		return nil, err
	}

	return res, nil
}

// hasPrefix checks if key has the given prefix
func hasPrefix(key, prefix []byte) bool {
	if len(key) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if key[i] != prefix[i] {
			return false
		}
	}
	return true
}

func (db *pebbleDB) Update(ctx context.Context, table string, key string, values map[string][]byte) error {
	// First read the existing data
	m, err := db.Read(ctx, table, key, nil)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("key not found: %s.%s", table, key)
	}

	// Update with new values
	for field, value := range values {
		m[field] = value
	}

	buf := db.bufPool.Get()
	defer db.bufPool.Put(buf)

	buf, err = db.r.Encode(buf, m)
	if err != nil {
		return err
	}

	rowKey := db.getRowKey(table, key)
	return db.db.Set(rowKey, buf, db.writeOpts)
}

func (db *pebbleDB) Insert(ctx context.Context, table string, key string, values map[string][]byte) error {
	rowKey := db.getRowKey(table, key)

	buf := db.bufPool.Get()
	defer db.bufPool.Put(buf)

	buf, err := db.r.Encode(buf, values)
	if err != nil {
		return err
	}

	return db.db.Set(rowKey, buf, db.writeOpts)
}

func (db *pebbleDB) Delete(ctx context.Context, table string, key string) error {
	rowKey := db.getRowKey(table, key)
	return db.db.Delete(rowKey, db.writeOpts)
}

func init() {
	ycsb.RegisterDBCreator("pebble", pebbleCreator{})
}
