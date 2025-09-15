package main

import (
	"log"
	"net/http"
	"os"
	"tenders/db"
	"tenders/db/migrations"
	"tenders/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	connString := os.Getenv("POSTGRES_CONN")
	if connString == "" {
		log.Fatal("POSTGRES_CONN env variable is not set")
	}

	dbConn, err := sqlx.Connect("postgres", connString)
	if err != nil {
		log.Fatalf("Cannot connect to DB: %v", err)
	}
	defer dbConn.Close()

	// Вызов функции миграций здесь
	// Пример если у вас миграции в пакете migrations с функцией Run()
	migrations.Run()

	store := db.NewStorage(dbConn)
	h := handlers.NewHandler(store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Get("/ping", h.PingHandler)
		// тендеры
		r.Post("/tenders/new", h.CreateTenderHandler)
		r.Get("/tenders", h.GetTendersHandler)
		r.Get("/tenders/my", h.GetUserTendersHandler)
		r.Put("/tenders/{tenderId}/status", h.ChangeTenderStatusHandler
		r.Put("/tenders/{tenderId}/rollback/{version}", h.RollbackTenderHandler)
		// предложения (bids)
		r.Post("/bids/new", h.CreateBidHandler)
		r.Get("/bids/my", h.GetUserBidsHandler)
		r.Put("/bids/{tenderId}/list", h.GetBidsForTenderHandler)
		r.Patch("/bids/{bidId}/edit", h.EditBidHandler)
		r.Put("/bids/{bidId}/status", h.UpdateBidStatusHandler)
		r.Put("/bids/{bidId}/rollback/{version}", h.RollbackBidHandler)
		r.Put("/bids/{bidId}/submit_decision", h.SubmitBidDecisionHandler)
		r.Get("bids/{tenderId}/reviews", h.GetBidReviewsHandler)
		r.Put("bids/{bidsId}/feedback", h.CreateBidFeedbackHandler)

		// Можно дальше добавить другие маршруты
	})

	serverAddr := os.Getenv("SERVER_ADDRESS")
	if serverAddr == "" {
		serverAddr = "0.0.0.0:8080"
	}

	log.Printf("Starting server on %s", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, r))
}
