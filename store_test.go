package main

import "testing"

var (
	db    *Store
	dir   string
	value = "a value"
)

func TestStore(t *testing.T) {
	dir = t.TempDir()
	db, err = NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}

	db.Set("somekey", value)

	val, err := db.Get("somekey")
	if err != nil {
		t.Fatal(err)
	}

	if val != value {
		t.Errorf("expected %q, got %q\n", value, val)
	}

	val, err = db.Get("fakekey")
	if err == nil {
		t.Errorf("expected %q, got %q\n", "error", val)
	}
}
