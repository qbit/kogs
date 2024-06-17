package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
)

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
	mux.HandleFunc("POST /users/create", makeCreate(reg, d))
	mux.HandleFunc("GET /users/auth", makeAuth(d))
	mux.HandleFunc("GET /syncs/progress/{document}", makeDocSync(d))
	mux.HandleFunc("PUT /syncs/progress", makeProgress(d))
	mux.HandleFunc("GET /healthcheck", healthHandler)
	mux.HandleFunc("/", slashHandler)

	s := http.Server{
		Handler: mux,
	}

	lis, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatal(err)
	}
	s.Serve(lis)
}
