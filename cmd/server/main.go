package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"tripshare/internal/database"
	"tripshare/internal/handler"
	"tripshare/internal/repository"
	"tripshare/internal/router"
	"tripshare/internal/service"
)

func main() {
	port := flag.String("port", "8080", "server port")
	dbPath := flag.String("db", "tripshare.db", "SQLite database path")
	flag.Parse()

	if p := os.Getenv("PORT"); p != "" {
		*port = p
	}
	if d := os.Getenv("DB_PATH"); d != "" {
		*dbPath = d
	}

	db := database.New(*dbPath)
	log.Printf("database initialized at %s", *dbPath)

	memberRepo := repository.NewMemberRepository(db)
	tripRepo := repository.NewTripRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	balanceSvc := service.NewBalanceService(memberRepo, tripRepo)
	memberSvc := service.NewMemberService(memberRepo, balanceSvc)
	tripSvc := service.NewTripService(tripRepo, memberRepo, balanceSvc)
	paymentSvc := service.NewPaymentService(paymentRepo, tripRepo, balanceSvc)

	memberH := handler.NewMemberHandler(memberSvc)
	tripH := handler.NewTripHandler(tripSvc)
	paymentH := handler.NewPaymentHandler(paymentSvc)
	balanceH := handler.NewBalanceHandler(balanceSvc)

	r := router.New(memberH, tripH, paymentH, balanceH)

	addr := fmt.Sprintf(":%s", *port)
	log.Printf("server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
