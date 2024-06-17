package main

import "testing"

func TestUserKey(t *testing.T) {
	u := &User{
		Username: "arst",
	}

	key := u.Key()
	exp := "user:arst:key"

	if key != exp {
		t.Errorf("expected %q, got %q\n", exp, key)
	}
}
