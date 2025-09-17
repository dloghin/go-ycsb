package leveldb

// LevelDB is a key-value storage library written at Google that provides an
// ordered mapping from string keys to string values. This package provides
// a Go implementation (syndtr/goleveldb) that can be used as a database backend
// for the YCSB benchmark.
//
// Configuration Properties:
//
// Basic options:
//   leveldb.dir - Database directory path (default: /tmp/leveldb)
//   leveldb.write_buffer - Write buffer size in bytes (default: 4MB)
//   leveldb.block_cache_capacity - Block cache capacity in bytes (default: 8MB)
//   leveldb.num_level - Number of levels (default: 7)
//   leveldb.compression_type - Compression type: "none" or "snappy" (default: "snappy")
//   leveldb.filter_policy - Filter policy: "bloom" for bloom filter (default: none)
//
// Advanced options:
//   leveldb.block_restart_interval - Block restart interval (default: 16)
//   leveldb.block_size - Block size in bytes (default: 4096)
//   leveldb.compaction_expand_limit - Compaction expand limit factor (default: 25)
//   leveldb.compaction_gp_overlaps - Compaction GP overlaps factor (default: 10)
//   leveldb.compaction_l0_trigger - L0 compaction trigger (default: 4)
//   leveldb.compaction_source_limit - Compaction source limit factor (default: 1)
//   leveldb.compaction_table_size - Compaction table size in bytes (default: 2MB)
//   leveldb.compaction_total_size - Compaction total size in bytes (default: 10MB)
//   leveldb.iterator_sampling_rate - Iterator sampling rate (default: 1MB)
//   leveldb.open_files_cache_capacity - Open files cache capacity (default: 500)
//
// Boolean options:
//   leveldb.disable_buffer_pool - Disable buffer pool (default: false)
//   leveldb.disable_block_cache - Disable block cache (default: false)
//   leveldb.disable_compaction_backoff - Disable compaction backoff (default: false)
//   leveldb.disable_large_batch_transaction - Disable large batch transaction (default: false)
//   leveldb.error_if_exist - Error if database exists (default: false)
//   leveldb.error_if_missing - Error if database is missing (default: false)
//   leveldb.no_sync - Disable sync writes (default: false)
//   leveldb.no_write_merge - Disable write merge (default: false)
//   leveldb.read_only - Open database in read-only mode (default: false)
//   leveldb.strict - Enable strict mode (default: false)
//
// Example usage:
//   go run cmd/go-ycsb/main.go run workloada -P workloads/workloada -p leveldb.dir=/tmp/test_leveldb
