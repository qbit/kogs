package main

import (
	"errors"
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
		log.Println(fmt.Errorf("failed to set %q: %s", key, err))
	}
}

func (s Store) Get(key string) (string, error) {
	keyPath := path.Join(string(s), key)
	_, err := os.Stat(keyPath)
	if errors.Is(err, os.ErrNotExist) {
		return "", os.ErrNotExist
	}

	data, err := os.ReadFile(path.Join(string(s), key))
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(data)), nil
}
