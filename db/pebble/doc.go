package pebble

// Pebble is a high-performance key-value storage engine developed by CockroachDB.
// It's designed as an evolution of LevelDB/RocksDB with better performance characteristics
// and is written in pure Go. This package provides a Pebble backend for the YCSB benchmark.
//
// Configuration Properties:
//
// Basic options:
//   pebble.dir - Database directory path (default: /tmp/pebble)
//   pebble.cache_size - Block cache size in bytes (default: 8MB)
//   pebble.bytes_per_sync - Bytes to sync periodically (default: 512KB)
//   pebble.disable_wal - Disable write-ahead log (default: false)
//   pebble.read_only - Open database in read-only mode (default: false)
//
// Error handling:
//   pebble.error_if_exists - Error if database exists (default: false)
//   pebble.error_if_not_exists - Error if database is missing (default: false)
//   pebble.error_if_not_pristine - Error if database is not pristine (default: false)
//
// Flush settings:
//   pebble.flush_delay_delete_range - Delay before flushing delete ranges in ms (default: 0)
//   pebble.flush_delay_range_key - Delay before flushing range keys in ms (default: 0)
//   pebble.flush_split_bytes - Target bytes per flush split interval (default: 0)
//
// L0 (Level 0) settings:
//   pebble.l0_compaction_file_threshold - L0 files to trigger compaction (default: 4)
//   pebble.l0_compaction_threshold - L0 read-amplification threshold (default: 4)
//   pebble.l0_stop_writes_threshold - L0 threshold to stop writes (default: 12)
//
// Base level settings:
//   pebble.lbase_max_bytes - Maximum bytes for base level (default: 64MB)
//
// File and memory settings:
//   pebble.max_manifest_file_size - Maximum manifest file size (default: 128MB)
//   pebble.max_open_files - Maximum open files (default: 1000)
//   pebble.memtable_size - MemTable size in bytes (default: 4MB)
//   pebble.memtable_stop_writes_threshold - MemTable threshold to stop writes (default: 2)
//
// Compaction settings:
//   pebble.max_concurrent_compactions - Maximum concurrent compactions (default: 1)
//   pebble.disable_automatic_compactions - Disable automatic compactions (default: false)
//
// Sync and cleanup settings:
//   pebble.no_sync_on_close - Disable sync on close (default: false)
//   pebble.num_prev_manifest - Number of previous manifests to keep (default: 1)
//   pebble.target_byte_deletion_rate - Rate limit for file deletions (default: 0)
//
// WAL (Write-Ahead Log) settings:
//   pebble.wal_bytes_per_sync - WAL bytes to sync (default: 0)
//   pebble.wal_dir - WAL directory path (default: same as database dir)
//   pebble.wal_min_sync_interval - Minimum WAL sync interval in microseconds (default: 0)
//
// Level options (applied to all levels):
//   pebble.level.block_restart_interval - Block restart interval (default: 16)
//   pebble.level.block_size - Block size in bytes (default: 4096)
//   pebble.level.block_size_threshold - Block size threshold percentage (default: 90)
//   pebble.level.compression - Compression type: "none", "snappy", "zstd" (default: "snappy")
//   pebble.level.filter_policy - Filter policy: "bloom" for bloom filter (default: none)
//   pebble.level.filter_type - Filter type: "table" or "block" (default: "table")
//   pebble.level.index_block_size - Index block size in bytes (default: 4096)
//   pebble.level.target_file_size - Target file size in bytes (default: 2MB)
//
// Performance Notes:
// - Pebble generally provides better performance than LevelDB, especially for write-heavy workloads
// - It has optimized compaction algorithms and better handling of large datasets
// - The pure Go implementation eliminates CGO overhead present in RocksDB bindings
// - Built-in range keys support and better iterator performance
//
// Example usage:
//   go run cmd/go-ycsb/main.go run workloada -P workloads/workloada -p pebble.dir=/tmp/test_pebble


