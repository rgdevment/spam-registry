# 🛡️ Global Spam Registry (GSR)

GSR is an open-source, high-performance backend system designed to identify, report, and score phone numbers based on spam/fraud risk.

## 🚀 Tech Stack
- **Language:** Go 1.23+
- **Database:** ScyllaDB (NoSQL)
- **Architecture:** Clean Architecture / Hexagonal
- **Pattern:** CQRS & Event Driven

## 📂 Project Structure
- `cmd/`: Entry points (API & Worker).
- `internal/domain/`: Core business logic & models.
- `internal/service/`: Business use cases.
- `internal/platform/`: Infrastructure implementations.

## 🛠️ Setup
```bash
make build
make run-api
