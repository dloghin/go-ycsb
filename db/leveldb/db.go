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

package leveldb

import (
	"context"
	"fmt"
	"os"

	"github.com/magiconair/properties"
	"github.com/pingcap/go-ycsb/pkg/prop"
	"github.com/pingcap/go-ycsb/pkg/util"
	"github.com/pingcap/go-ycsb/pkg/ycsb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

// properties
const (
	leveldbDir                          = "leveldb.dir"
	leveldbWriteBuffer                  = "leveldb.write_buffer"
	leveldbBlockCacheCapacity           = "leveldb.block_cache_capacity"
	leveldbBlockRestartInterval         = "leveldb.block_restart_interval"
	leveldbBlockSize                    = "leveldb.block_size"
	leveldbCompactionExpandLimit        = "leveldb.compaction_expand_limit"
	leveldbCompactionGPOverlaps         = "leveldb.compaction_gp_overlaps"
	leveldbCompactionL0Trigger          = "leveldb.compaction_l0_trigger"
	leveldbCompactionSourceLimit        = "leveldb.compaction_source_limit"
	leveldbCompactionTableSize          = "leveldb.compaction_table_size"
	leveldbCompactionTotalSize          = "leveldb.compaction_total_size"
	leveldbCompressionType              = "leveldb.compression_type"
	leveldbIteratorSamplingRate         = "leveldb.iterator_sampling_rate"
	leveldbNumLevel                     = "leveldb.num_level"
	leveldbOpenFilesCacheCapacity       = "leveldb.open_files_cache_capacity"
	leveldbFilterPolicy                 = "leveldb.filter_policy"
	leveldbDisableBufferPool            = "leveldb.disable_buffer_pool"
	leveldbDisableBlockCache            = "leveldb.disable_block_cache"
	leveldbDisableCompactionBackoff     = "leveldb.disable_compaction_backoff"
	leveldbDisableLargeBatchTransaction = "leveldb.disable_large_batch_transaction"
	leveldbErrorIfExist                 = "leveldb.error_if_exist"
	leveldbErrorIfMissing               = "leveldb.error_if_missing"
	leveldbNoSync                       = "leveldb.no_sync"
	leveldbNoWriteMerge                 = "leveldb.no_write_merge"
	leveldbReadOnly                     = "leveldb.read_only"
	leveldbStrict                       = "leveldb.strict"
)

type levelDBCreator struct{}

type levelDBDB struct {
	p *properties.Properties

	db *leveldb.DB

	r       *util.RowCodec
	bufPool *util.BufPool

	readOpts  *opt.ReadOptions
	writeOpts *opt.WriteOptions
}

func (c levelDBCreator) Create(p *properties.Properties) (ycsb.DB, error) {
	dir := p.GetString(leveldbDir, "/tmp/leveldb")

	if p.GetBool(prop.DropData, prop.DropDataDefault) {
		os.RemoveAll(dir)
	}

	opts := getOptions(p)

	db, err := leveldb.OpenFile(dir, opts)
	if err != nil {
		return nil, err
	}

	return &levelDBDB{
		p:         p,
		db:        db,
		r:         util.NewRowCodec(p),
		bufPool:   util.NewBufPool(),
		readOpts:  getReadOptions(p),
		writeOpts: getWriteOptions(p),
	}, nil
}

func getOptions(p *properties.Properties) *opt.Options {
	opts := &opt.Options{}

	// Database options
	opts.WriteBuffer = p.GetInt(leveldbWriteBuffer, 4*1024*1024)
	opts.BlockCacheCapacity = p.GetInt(leveldbBlockCacheCapacity, 8*1024*1024)
	opts.BlockRestartInterval = p.GetInt(leveldbBlockRestartInterval, 16)
	opts.BlockSize = p.GetInt(leveldbBlockSize, 4096)
	opts.CompactionExpandLimitFactor = p.GetInt(leveldbCompactionExpandLimit, 25)
	opts.CompactionGPOverlapsFactor = p.GetInt(leveldbCompactionGPOverlaps, 10)
	opts.CompactionL0Trigger = p.GetInt(leveldbCompactionL0Trigger, 4)
	opts.CompactionSourceLimitFactor = p.GetInt(leveldbCompactionSourceLimit, 1)
	opts.CompactionTableSize = p.GetInt(leveldbCompactionTableSize, 2*1024*1024)
	opts.CompactionTotalSize = p.GetInt(leveldbCompactionTotalSize, 10*1024*1024)
	opts.IteratorSamplingRate = p.GetInt(leveldbIteratorSamplingRate, 1*1024*1024)
	// NumLevel is not configurable in goleveldb, using default
	opts.OpenFilesCacheCapacity = p.GetInt(leveldbOpenFilesCacheCapacity, 500)

	// Compression type
	compressionType := p.GetString(leveldbCompressionType, "snappy")
	switch compressionType {
	case "none":
		opts.Compression = opt.NoCompression
	case "snappy":
		opts.Compression = opt.SnappyCompression
	default:
		opts.Compression = opt.SnappyCompression
	}

	// Filter policy
	if filterPolicy := p.GetString(leveldbFilterPolicy, ""); filterPolicy == "bloom" {
		opts.Filter = filter.NewBloomFilter(10) // 10 bits per key
	}

	// Boolean options
	opts.DisableBufferPool = p.GetBool(leveldbDisableBufferPool, false)
	opts.DisableBlockCache = p.GetBool(leveldbDisableBlockCache, false)
	opts.DisableCompactionBackoff = p.GetBool(leveldbDisableCompactionBackoff, false)
	opts.DisableLargeBatchTransaction = p.GetBool(leveldbDisableLargeBatchTransaction, false)
	opts.ErrorIfExist = p.GetBool(leveldbErrorIfExist, false)
	opts.ErrorIfMissing = p.GetBool(leveldbErrorIfMissing, false)
	opts.NoSync = p.GetBool(leveldbNoSync, false)
	opts.NoWriteMerge = p.GetBool(leveldbNoWriteMerge, false)
	opts.ReadOnly = p.GetBool(leveldbReadOnly, false)

	// Strict mode options
	if p.GetBool(leveldbStrict, false) {
		opts.Strict = opt.StrictAll
	} else {
		opts.Strict = opt.DefaultStrict
	}

	return opts
}

func getReadOptions(p *properties.Properties) *opt.ReadOptions {
	return &opt.ReadOptions{}
}

func getWriteOptions(p *properties.Properties) *opt.WriteOptions {
	opts := &opt.WriteOptions{}
	opts.NoWriteMerge = p.GetBool(leveldbNoWriteMerge, false)
	opts.Sync = !p.GetBool(leveldbNoSync, false)
	return opts
}

func (db *levelDBDB) Close() error {
	return db.db.Close()
}

func (db *levelDBDB) InitThread(ctx context.Context, _ int, _ int) context.Context {
	return ctx
}

func (db *levelDBDB) CleanupThread(_ context.Context) {
}

func (db *levelDBDB) getRowKey(table string, key string) []byte {
	return util.Slice(fmt.Sprintf("%s:%s", table, key))
}

func (db *levelDBDB) Read(ctx context.Context, table string, key string, fields []string) (map[string][]byte, error) {
	value, err := db.db.Get(db.getRowKey(table, key), db.readOpts)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return db.r.Decode(value, fields)
}

func (db *levelDBDB) Scan(ctx context.Context, table string, startKey string, count int, fields []string) ([]map[string][]byte, error) {
	res := make([]map[string][]byte, 0, count)

	iter := db.db.NewIterator(nil, db.readOpts)
	defer iter.Release()

	rowStartKey := db.getRowKey(table, startKey)
	tablePrefix := []byte(table + ":")

	// Start from the specified key
	if !iter.Seek(rowStartKey) {
		return res, iter.Error()
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

func (db *levelDBDB) Update(ctx context.Context, table string, key string, values map[string][]byte) error {
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
	return db.db.Put(rowKey, buf, db.writeOpts)
}

func (db *levelDBDB) Insert(ctx context.Context, table string, key string, values map[string][]byte) error {
	rowKey := db.getRowKey(table, key)

	buf := db.bufPool.Get()
	defer db.bufPool.Put(buf)

	buf, err := db.r.Encode(buf, values)
	if err != nil {
		return err
	}

	return db.db.Put(rowKey, buf, db.writeOpts)
}

func (db *levelDBDB) Delete(ctx context.Context, table string, key string) error {
	rowKey := db.getRowKey(table, key)
	return db.db.Delete(rowKey, db.writeOpts)
}

func init() {
	ycsb.RegisterDBCreator("leveldb", levelDBCreator{})
}
