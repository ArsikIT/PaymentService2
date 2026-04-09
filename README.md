# Payment Service

`payment-service` is the payment bounded context of the assignment platform. It owns only payment data and does not know about order persistence. Its responsibility is to authorize or decline payments and store the transaction result.

## Architecture

Project layout:

```text
payment-service/
|- cmd/main.go
|- internal/domain
|- internal/usecase
|- internal/repository
|- internal/transport/http
|- internal/app
|- migrations
|- docker-compose.yml
`- README.md
```

Dependency direction:

```text
HTTP handlers -> use cases -> repository port
                              `-> postgres repository
```

Key decisions:

- `domain` contains only payment entities, statuses, and business errors.
- `usecase` contains payment validation, limit checks, status selection, and ID generation.
- `repository` contains PostgreSQL persistence only.
- `transport/http` stays thin and only parses requests and returns DTO responses.
- `cmd/main.go` and `internal/app/server.go` act as the composition root and manual dependency wiring.

## Bounded Context

`payment-service` owns:

- payment authorization result
- transaction identifier
- payment persistence

`payment-service` does not own:

- order lifecycle
- order cancellation rules
- order persistence

This separation is important for defense: the service stores payment data only and is called by `order-service` through REST.

## Business Rules

- Money uses `int64` cents only.
- `order_id` must be provided.
- `amount` must be greater than `0`.
- If `amount <= 100000`, payment becomes `Authorized` and gets a unique `transaction_id`.
- If `amount > 100000`, payment becomes `Declined`.
- Every payment attempt is stored in the payment database.

## Database Per Service

This service has its own PostgreSQL container and its own database:

- container: `payment-db`
- database: `payment_service`
- port: `55432`

`payment-service` does not read or write order tables.

## Run

1. Start the payment database:

```bash
docker compose up -d
```

2. Run the service:

```bash
go run ./cmd
```

## Environment Variables

- `HTTP_ADDR` default: `:8081`
- `POSTGRES_DSN` default: `postgres://postgres@127.0.0.1:55432/payment_service?sslmode=disable`

## API Examples

Create payment:

```bash
curl -X POST http://localhost:8081/payments \
  -H "Content-Type: application/json" \
  -d "{\"order_id\":\"ord-1\",\"amount\":15000}"
```

Response example:

```json
{
  "id": "pay-123",
  "order_id": "ord-1",
  "transaction_id": "tx-123",
  "amount": 15000,
  "status": "Authorized"
}
```

Get payment by order id:

```bash
curl http://localhost:8081/payments/ord-1
```

Declined payment example:

```bash
curl -X POST http://localhost:8081/payments \
  -H "Content-Type: application/json" \
  -d "{\"order_id\":\"ord-2\",\"amount\":150001}"
```

## Architecture Diagram

```mermaid
flowchart LR
    Client --> PaymentAPI[Payment Service HTTP API]
    PaymentAPI --> PaymentUC[Payment Use Case]
    PaymentUC --> PaymentRepo[(Payment Service DB)]
```
