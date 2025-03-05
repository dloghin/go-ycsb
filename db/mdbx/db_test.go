package mdbx

import (
	"context"
	"testing"

	"github.com/magiconair/properties"
)

func TestMdbxOps(t *testing.T) {
	props := properties.NewProperties()
	mdbxCreator := mdbxCreator{}
	db, err := mdbxCreator.Create(props)
	if err != nil {
		t.Errorf("Error creating db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	err = db.Insert(ctx, "table", "key1", map[string][]byte{"field1": []byte("value11"), "field2": []byte("value12")})
	if err != nil {
		t.Errorf("Error inserting data: %v", err)
	}

	err = db.Insert(ctx, "table", "key2", map[string][]byte{"field1": []byte("value21"), "field2": []byte("value22")})
	if err != nil {
		t.Errorf("Error inserting data: %v", err)
	}

	res, err := db.Read(ctx, "table", "key1", []string{"field1", "field2"})
	if err != nil {
		t.Errorf("Error reading data: %v", err)
	}
	if string(res["field1"]) != "value11" || string(res["field2"]) != "value12" {
		t.Errorf("Error reading data: %v", res)
	}
}

func TestMdbxMultiDB(t *testing.T) {
	props1 := properties.NewProperties()
	mdbxCreator := mdbxCreator{}
	db1, err := mdbxCreator.Create(props1)
	if err != nil {
		t.Errorf("Error creating db1: %v", err)
	}
	defer db1.Close()

	props2 := properties.NewProperties()
	props2.Set("mdbx.path", "test2")
	props2.Set("mdbx.table", "table2")
	db2, err := mdbxCreator.Create(props2)
	if err != nil {
		t.Errorf("Error creating db2: %v", err)
	}
	defer db2.Close()

	ctx := context.Background()
	err = db1.Insert(ctx, "table", "key1", map[string][]byte{"field1": []byte("value11"), "field2": []byte("value12")})
	if err != nil {
		t.Errorf("Error inserting data: %v", err)
	}

	err = db2.Insert(ctx, "table", "key2", map[string][]byte{"field1": []byte("value21"), "field2": []byte("value22")})
	if err != nil {
		t.Errorf("Error inserting data: %v", err)
	}

	res1, err := db1.Read(ctx, "table", "key1", []string{"field1", "field2"})
	if err != nil {
		t.Errorf("Error reading data: %v", err)
	}
	if string(res1["field1"]) != "value11" || string(res1["field2"]) != "value12" {
		t.Errorf("Invalid data read from db: %v", res1)
	}

	res2, err := db2.Read(ctx, "table", "key2", []string{"field1", "field2"})
	if err != nil {
		t.Errorf("Error reading data: %v", err)
	}
	if string(res2["field1"]) != "value21" || string(res2["field2"]) != "value22" {
		t.Errorf("Invalid data read from db: %v", res2)
	}
}
