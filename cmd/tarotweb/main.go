package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"web-app-agent-runtime/internal/tarotweb"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	app, err := tarotweb.New()
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              *addr,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Printf("tarot web app listening on http://localhost%s\n", *addr)
	log.Fatal(server.ListenAndServe())
}
