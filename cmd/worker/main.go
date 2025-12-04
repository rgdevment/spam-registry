package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/rgdevment/spam-registry/internal/platform/storage/scylla"
	"github.com/rgdevment/spam-registry/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, using system environment variables")
	}

	phonePtr := flag.String("phone", "", "The phone number to recalculate risk for (E.164 format)")
	flag.Parse()

	if *phonePtr == "" {
		log.Fatal("❌ Error: You must provide a phone number.\nUsage: go run cmd/worker/main.go -phone=+56912345678")
	}

	log.Printf("🐝 GSR Worker Starting manually for target: %s", *phonePtr)

	scyllaHost := os.Getenv("SCYLLA_HOST")
	keyspace := os.Getenv("SCYLLA_KEYSPACE")
	if scyllaHost == "" {
		scyllaHost = "localhost"
	}

	session, err := scylla.Connect(keyspace, scyllaHost)
	if err != nil {
		log.Fatalf("❌ DB Connection Failed: %v", err)
	}
	defer session.Close()

	repo := scylla.NewScyllaRepository(session)

	svc := service.NewReportService(repo, "")

	log.Println("🧠 Running Quantum Algorithm...")
	err = svc.CalculateAndSaveRisk(context.Background(), *phonePtr)
	if err != nil {
		log.Fatalf("❌ Calculation Failed: %v", err)
	}

	log.Println("✅ Success! Score updated in ScyllaDB (Scores & Active Threats tables).")
}
