package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func NewStore(s string) (*Store, error) {
	fi, err := os.Lstat(s)
	if err != nil {
		return nil, err
	}

	if !fi.IsDir() {
		return nil, fmt.Errorf("not a directory")
	}
	fstore := Store(s)
	return &fstore, nil
}

type Store string

func (s Store) Set(key string, value string) {
	err := os.WriteFile(path.Join(string(s), key), []byte(value), 0600)
	if err != nil {
		log.Println(err)
	}
}

func (s Store) Get(key string) (string, error) {
	data, err := os.ReadFile(path.Join(string(s), key))
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(data)), nil
}
