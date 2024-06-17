package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func httpLog(r *http.Request) {
	n := time.Now()
	fmt.Printf("%s (%s) [%s] \"%s %s\" %03d\n",
		r.RemoteAddr,
		n.Format(time.RFC822Z),
		r.Method,
		r.URL.Path,
		r.Proto,
		r.ContentLength,
	)
}

func makeAuth(d *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		httpLog(r)
		_, err := authUserFromHeader(d, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Unauthorized"}`))
			return
		}
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"authorized": "OK"}`))
	}
}

func makeProgress(d *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		httpLog(r)
		u, err := authUserFromHeader(d, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Unauthorized"}`))
			return
		}
		prog := Progress{}
		dec := json.NewDecoder(r.Body)
		err = dec.Decode(&prog)
		if err != nil {
			log.Println(err)
			http.Error(w, "invalid document", http.StatusNotFound)
			return
		}

		prog.User = *u
		prog.Save(d)

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`{"document": "%s", "timestamp": "%d"}`, prog.Document, prog.Timestamp)))
	}
}

func makeCreate(reg *bool, d *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		httpLog(r)
		if !*reg {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Registration disabled"}`))
			return
		}
		u := User{}

		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&u)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		_, err = d.Get(u.Key())
		if err != nil {
			d.Set(u.Key(), u.Password)
		} else {
			log.Println(err)
			http.Error(w, "Username is already registered", http.StatusPaymentRequired)
			return
		}

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(201)
		w.Write(u.Created())
	}
}

func makeDocSync(d *Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		httpLog(r)

		// TODO: I have no idea why this PathValue returns "".. dirty hack
		// to grab it from the URL anyway :(
		doc := r.PathValue("document")
		if doc == "" {
			parts := strings.Split(r.URL.String(), "/")
			doc = parts[len(parts)-1]
			if doc == "" {
				http.Error(w, "Invalid Request", http.StatusBadRequest)
				return
			}
		}

		u, err := authUserFromHeader(d, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Unauthorized"}`))
			return
		}
		prog := Progress{
			Document: doc,
			User:     *u,
		}

		err = prog.Get(d)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(prog)
		if err != nil {
			log.Println(err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(200)
		w.Write(b)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	httpLog(r)
	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(`{"state": "OK"}`))
}

func slashHandler(w http.ResponseWriter, r *http.Request) {
	httpLog(r)
	w.Header().Add("Content-type", "text/plain")
	w.WriteHeader(200)
	w.Write([]byte(`kogs: koreader sync server`))
}
