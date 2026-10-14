package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "modernc.org/sqlite"
)

func main() {
	addr := flag.String("addr", ":8080", "listen addr")
	dbPath := flag.String("db", "./data/rfi.db", "sqlite file path")
	dataDir := flag.String("data", "./data", "data dir (screenshots/, models/)")
	frontendOrigin := flag.String("cors", "http://localhost:5173", "允許的前端 origin（Vite 預設）")
	flag.Parse()

	db, err := OpenDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	srv := NewServer(db, *dataDir)
	mux := http.NewServeMux()
	srv.routes(mux)

	handler := withCORS(mux, *frontendOrigin)
	fmt.Printf("RFI board listening on %s (db=%s)\n", *addr, *dbPath)
	log.Fatal(http.ListenAndServe(*addr, handler))
}

func withCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
