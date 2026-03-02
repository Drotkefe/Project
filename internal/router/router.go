package router

import (
	"log"
	"net/http"
	"time"
	"tripshare/internal/handler"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func New(
	memberH *handler.MemberHandler,
	tripH *handler.TripHandler,
	paymentH *handler.PaymentHandler,
	balanceH *handler.BalanceHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Members
	mux.HandleFunc("POST /members", memberH.Create)
	mux.HandleFunc("GET /members", memberH.List)
	mux.HandleFunc("PUT /members/{id}", memberH.Update)
	mux.HandleFunc("DELETE /members/{id}", memberH.Delete)

	// Trips
	mux.HandleFunc("POST /trips", tripH.Create)
	mux.HandleFunc("GET /trips", tripH.List)
	mux.HandleFunc("GET /trips/{id}", tripH.Get)
	mux.HandleFunc("PUT /trips/{id}", tripH.Update)
	mux.HandleFunc("DELETE /trips/{id}", tripH.Delete)

	// Payments
	mux.HandleFunc("POST /trips/{id}/payments", paymentH.Create)
	mux.HandleFunc("PUT /trips/{id}/payments/{paymentId}", paymentH.Update)
	mux.HandleFunc("DELETE /trips/{id}/payments/{paymentId}", paymentH.Delete)

	// Balances
	mux.HandleFunc("GET /balances", balanceH.Get)

	// Static files and UI (no-cache so updates are picked up immediately)
	staticFS := http.StripPrefix("/static/", http.FileServer(http.Dir("web/static")))
	mux.Handle("GET /static/", noCacheMiddleware(staticFS))
	mux.HandleFunc("GET /", serveIndex)

	return loggingMiddleware(mux)
}

func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "web/templates/index.html")
}
