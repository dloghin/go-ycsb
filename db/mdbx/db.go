package mdbx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/c2h5oh/datasize"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
	logv3 "github.com/ledgerwatch/log/v3"
	"github.com/magiconair/properties"
	"github.com/pingcap/go-ycsb/pkg/util"
	"github.com/pingcap/go-ycsb/pkg/ycsb"
)

type mdbxCreator struct{}

// mdbxDB implements the ycsb.DB interface.
type mdbxDB struct {
	db      kv.RwDB
	p       *properties.Properties
	r       *util.RowCodec
	table   string
	bufPool *util.BufPool
}

func (c mdbxCreator) Create(p *properties.Properties) (ycsb.DB, error) {
	fmt.Println("mdbxDB: Create() called")

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path := filepath.Dir(ex)
	logger := logv3.New()
	table := "testtable"
	m := mdbx.NewMDBX(logger).InMem(path).WithTableCfg(func(defaultBuckets kv.TableCfg) kv.TableCfg {
		return kv.TableCfg{
			table:       kv.TableCfgItem{Flags: kv.DupSort},
			kv.Sequence: kv.TableCfgItem{},
		}
	}).MapSize(8192 * datasize.MB).MustOpen()

	ret := &mdbxDB{
		db:      m,
		table:   table,
		p:       p,
		r:       util.NewRowCodec(p),
		bufPool: util.NewBufPool(),
	}
	return ret, nil
}

func (db *mdbxDB) Close() error {
	fmt.Println("mdbxDB: Close() called")
	db.db.Close()
	return nil
}

func (db *mdbxDB) InitThread(ctx context.Context, _ int, _ int) context.Context {
	fmt.Println("mdbxDB: InitThread() called")
	return ctx
}

func (db *mdbxDB) CleanupThread(_ context.Context) {
	fmt.Println("mdbxDB: CleanupThread() called")
}

func (db *mdbxDB) Read(ctx context.Context, table string, key string, fields []string) (map[string][]byte, error) {
	tx, err := db.db.BeginRo(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	val, err := tx.GetOne(db.table, []byte(key))
	if err != nil {
		return nil, err
	}
	return db.r.Decode(val, fields)
}

func (db *mdbxDB) Scan(ctx context.Context, table string, startKey string, count int, fields []string) ([]map[string][]byte, error) {
	return nil, fmt.Errorf("mdbxDB: Scan() not implemented")
}

func (db *mdbxDB) Update(ctx context.Context, table string, key string, values map[string][]byte) error {
	return db.Insert(ctx, table, key, values)
}

func (db *mdbxDB) Insert(ctx context.Context, table string, key string, values map[string][]byte) error {
	tx, err := db.db.BeginRw(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	buf := db.bufPool.Get()
	defer db.bufPool.Put(buf)
	buf, err = db.r.Encode(buf, values)
	if err != nil {
		return err
	}
	err = tx.Put(db.table, []byte(key), buf)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *mdbxDB) Delete(ctx context.Context, table string, key string) error {
	tx, err := db.db.BeginRw(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	err = tx.Delete(db.table, []byte(key))
	if err != nil {
		return err
	}
	return tx.Commit()
}

func init() {
	ycsb.RegisterDBCreator("mdbx", mdbxCreator{})
}
