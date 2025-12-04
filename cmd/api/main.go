package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	middleware "github.com/rgdevment/spam-registry/internal/platform/http/middleware"

	httpHandler "github.com/rgdevment/spam-registry/internal/platform/http"
	"github.com/rgdevment/spam-registry/internal/platform/storage/scylla"
	"github.com/rgdevment/spam-registry/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	apiKey := os.Getenv("API_MASTER_KEY")
	if apiKey == "" {
		log.Fatal("❌ API_MASTER_KEY is required in .env")
	}

	scyllaHost := os.Getenv("SCYLLA_HOST")
	keyspace := os.Getenv("SCYLLA_KEYSPACE")
	saltSecret := os.Getenv("APP_SALT_SECRET")
	port := os.Getenv("HTTP_PORT")

	if scyllaHost == "" {
		scyllaHost = "localhost"
	}
	if port == "" {
		port = ":8080"
	}

	log.Println("🛡️  Iniciando Global Spam Registry (GSR)...")

	session, err := scylla.Connect(keyspace, scyllaHost)
	if err != nil {
		log.Fatalf("❌ Error conectando a ScyllaDB: %v", err)
	}
	defer session.Close()

	repo := scylla.NewScyllaRepository(session)

	svc := service.NewReportService(repo, saltSecret)

	handler := httpHandler.NewHandler(svc)

	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.APIKeyAuth(apiKey))

	handler.RegisterRoutes(r)

	log.Printf("🚀 Servidor escuchando en http://localhost%s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatalf("❌ Error en el servidor HTTP: %v", err)
	}
}
