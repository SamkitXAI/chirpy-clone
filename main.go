package main

import (
	"log"
	"net/http"
	"sync/atomic"
	"strconv"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	count := cfg.fileserverHits.Load()
	w.Write([]byte("Hits: " + strconv.Itoa(int(count))))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("Hits reset to 0"))
}


func main() {
	// serve files out of the current folder
	const filepathRoot = "."

	// pick a port and stick to it
	const port = "8080"

	apiCfg := apiConfig{}

	// set up a new router
	mux := http.NewServeMux()

	// create the file server rooted at "."
	fs := http.FileServer(http.Dir(filepathRoot))
	// strip off "/app/" before handing it to fs
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", fs)))

	mux.HandleFunc("GET /metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /reset", apiCfg.handlerReset)


	// bundle up our server settings
	srv := &http.Server{
		Addr:    ":" + port, // listen on this port
		Handler: mux,        // use our mux for routing
	}

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		// set content-type
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		// send a 200 OK
		w.WriteHeader(http.StatusOK)
		// write the body
		w.Write([]byte("OK"))
	})

	// let us know where we're serving from
	log.Printf("Serving files from %s on port %s", filepathRoot, port)

	// start the show; crash hard if something goes wrong
	log.Fatal(srv.ListenAndServe())
}
