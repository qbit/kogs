package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	testDB   *Store
	err      error
	token    string
	user     = "arst"
	password = "arstarst"
	document = "arstarstarstarst"
	usrJson  = []byte(
		fmt.Sprintf(`{"username": "%s", "password": "%s"}`, user, password),
	)
	progressJson = []byte(
		fmt.Sprintf(`{"device": "snake", "progress": "30", "document": "%s", "percentage": 0.1, "device_id": "1234", "timestamp": 1711992660}`, document),
	)
)

func TestMain(t *testing.M) {
	os.RemoveAll("./test_db")
	os.MkdirAll("./test_db", 0755)

	t.Run()

	os.RemoveAll("./test_db")
}

func TestCreate(t *testing.T) {
	req, err := http.NewRequest("POST", "/users/create", bytes.NewBuffer(usrJson))
	if err != nil {
		t.Fatal(err)
	}

	db, err := NewStore("./test_db")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	allow := true
	handler := http.HandlerFunc(makeCreate(&allow, db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("expected %v, got %v\n", http.StatusCreated, status)
	}

	u := &User{Username: user, Password: password}
	token, err = db.Get(u.Key())
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatalf("token is empty, %q", token)
	}
}

func TestCreateDup(t *testing.T) {
	req, err := http.NewRequest("POST", "/users/create", bytes.NewBuffer(usrJson))
	if err != nil {
		t.Fatal(err)
	}

	db, err := NewStore("./test_db")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	allow := true
	handler := http.HandlerFunc(makeCreate(&allow, db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusPaymentRequired {
		t.Errorf("expected %v, got %v\n", http.StatusPaymentRequired, status)
	}
}

func TestAuth(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/auth", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("x-auth-user", user)
	req.Header.Set("x-auth-key", token)

	db, err := NewStore("./test_db")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeAuth(db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected %v, got %v\n", http.StatusOK, status)
	}
}

func TestAuthDenied(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/auth", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("x-auth-user", user)
	req.Header.Set("x-auth-key", fmt.Sprintf("bad_%s", token))

	db, err := NewStore("./test_db")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeAuth(db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected %v, got %v\n", http.StatusUnauthorized, status)
	}
}

func TestProgress(t *testing.T) {
	req, err := http.NewRequest("PUT", "/syncs/progress", bytes.NewBuffer(progressJson))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("x-auth-user", user)
	req.Header.Set("x-auth-key", token)

	db, err := NewStore("./test_db")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeProgress(db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected %v, got %v\n", http.StatusOK, status)
	}
}

func TestProgressDenied(t *testing.T) {
	req, err := http.NewRequest("PUT", "/syncs/progress", bytes.NewBuffer(progressJson))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("x-auth-user", user)
	req.Header.Set("x-auth-key", fmt.Sprintf("%s_bad", token))

	db, err := NewStore("./test_db")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeProgress(db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected %v, got %v\n", http.StatusUnauthorized, status)
	}
}

func TestGetProgress(t *testing.T) {
	docURL := fmt.Sprintf("/sync/progress/%s", document)
	req, err := http.NewRequest("GET", docURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("x-auth-user", user)
	req.Header.Set("x-auth-key", token)

	db, err := NewStore("./test_db")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeDocSync(db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected %v, got %v\n", http.StatusOK, status)
	}

	prog := &Progress{}
	dec := json.NewDecoder(rr.Body)
	err = dec.Decode(prog)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetProgressDenied(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/syncs/progress/%s", document), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("x-auth-user", user)
	req.Header.Set("x-auth-key", fmt.Sprintf("%s_bad", token))

	db, err := NewStore("./test_db")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeDocSync(db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected %v, got %v\n", http.StatusUnauthorized, status)
	}
}

func TestGetInvalidDoc(t *testing.T) {
	req, err := http.NewRequest("GET", fmt.Sprintf("/syncs/progress/%s_fake", document), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("x-auth-user", user)
	req.Header.Set("x-auth-key", token)

	db, err := NewStore("./test_db")
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(makeDocSync(db))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("expected %v, got %v\n", http.StatusInternalServerError, status)
	}
}

func TestHealthCheck(t *testing.T) {
	req, err := http.NewRequest("GET", "/healthcheck", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected %v, got %v\n", http.StatusOK, status)
	}
}
