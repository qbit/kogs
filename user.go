package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct {
	Username string `json:"username"`
	Password string
	AuthKey  string
}

func (u *User) Key() string {
	return fmt.Sprintf("user:%s:key", u.Username)
}

func (u *User) Auth(authKey string) bool {
	return u.AuthKey == authKey
}

func (u *User) Created() []byte {
	j, _ := json.Marshal(u)
	return j
}

func authUserFromHeader(d *Store, r *http.Request) (*User, error) {
	un := r.Header.Get("x-auth-user")
	uk := r.Header.Get("x-auth-key")

	u := &User{
		Username: un,
	}
	storedKey, err := d.Get(u.Key())
	if err != nil {
		// No user
		return nil, err
	}

	u.AuthKey = string(storedKey)
	if u.Auth(uk) {
		return u, nil
	}

	return nil, fmt.Errorf("Unauthorized")
}
