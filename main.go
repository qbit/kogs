package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
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

type Progress struct {
	Device     string  `json:"device"`
	Progress   string  `json:"progress"`
	Document   string  `json:"document"`
	Percentage float64 `json:"percentage"`
	DeviceID   string  `json:"device_id"`
	Timestamp  int64   `json:"timestamp"`
	User       User
}

func (p *Progress) DocKey() string {
	return fmt.Sprintf("user:%s:document:%s", p.User.Username, p.Document)
}

func (p *Progress) Save(d *Store) {
	d.Set(p.DocKey()+"_percent", fmt.Sprintf("%f", p.Percentage))
	d.Set(p.DocKey()+"_progress", p.Progress)
	d.Set(p.DocKey()+"_device", p.Device)
	d.Set(p.DocKey()+"_device_id", p.DeviceID)
	d.Set(p.DocKey()+"_timestamp", fmt.Sprintf("%d", (time.Now().Unix())))
}

func (p *Progress) Get(d *Store) error {
	if p.Document == "" {
		return fmt.Errorf("invalid document")
	}
	pct, err := d.Get(p.DocKey() + "_percent")
	if err != nil {
		return err
	}
	p.Percentage, _ = strconv.ParseFloat(string(pct), 64)

	prog, err := d.Get(p.DocKey() + "_progress")
	if err != nil {
		return err
	}
	p.Progress = string(prog)

	dev, err := d.Get(p.DocKey() + "_device")
	if err != nil {
		return err
	}
	p.Device = string(dev)

	devID, err := d.Get(p.DocKey() + "_device_id")
	if err != nil {
		return err
	}
	p.DeviceID = string(devID)

	ts, err := d.Get(p.DocKey() + "_timestamp")
	if err != nil {
		return err
	}
	stamp, err := strconv.ParseInt(string(ts), 10, 64)
	if err != nil {
		return err
	}

	p.Timestamp = stamp

	return nil
}

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

func main() {
	reg := flag.Bool("reg", true, "enable user registration")
	listen := flag.String("listen", ":8383", "interface and port to listen on")
	dbDir := flag.String("db", "db", "full path to database directory")
	flag.Parse()

	err := os.MkdirAll(*dbDir, 0750)
	if err != nil {
		log.Fatal(err)
	}

	d, err := NewStore(*dbDir)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Storing data in: %q", *dbDir)

	if !*reg {
		log.Println("registration disabled")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users/create", func(w http.ResponseWriter, r *http.Request) {
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
			log.Println("user exists", err)
			http.Error(w, "Username is already registered", http.StatusPaymentRequired)
			return
		}

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(201)
		w.Write(u.Created())
	})
	mux.HandleFunc("GET /users/auth", func(w http.ResponseWriter, r *http.Request) {
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
	})

	mux.HandleFunc("PUT /syncs/progress", func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		prog.User = *u
		prog.Save(d)

		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(fmt.Sprintf(`{"document": "%s", "timestamp": "%d"}`, prog.Document, prog.Timestamp)))
	})
	mux.HandleFunc("GET /syncs/progress/{document}", func(w http.ResponseWriter, r *http.Request) {
		httpLog(r)
		u, err := authUserFromHeader(d, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message": "Unauthorized"}`))
			return
		}
		prog := Progress{
			Document: r.PathValue("document"),
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
	})

	mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, r *http.Request) {
		httpLog(r)
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"state": "OK"}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpLog(r)
		w.Header().Add("Content-type", "text/plain")
		w.WriteHeader(200)
		w.Write([]byte(`kogs: koreader sync server`))
	})

	s := http.Server{
		Handler: mux,
	}

	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	s.Serve(lis)
}
