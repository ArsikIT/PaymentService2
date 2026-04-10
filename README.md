# Payment Service

`payment-service` owns payment data, exposes a REST API for manual checks, and exposes a gRPC server used by `order-service` to authorize payments.

## Architecture

Project layout:

```text
payment-service/
|- cmd/main.go
|- internal/domain
|- internal/usecase
|- internal/repository
|- internal/transport/http
|- internal/transport/grpc
|- internal/app
|- migrations
|- docker-compose.yml
`- README.md
```

Dependency direction:

```text
HTTP handlers -> use cases -> repository port
gRPC handlers -> use cases -> repository port
                               `-> postgres repository
```

Key decisions:

- `domain` contains only payment entities, statuses, and business errors.
- `usecase` contains payment validation, limit checks, status selection, and ID generation.
- `repository` contains PostgreSQL persistence only.
- `transport/http` stays thin and maps REST requests/responses.
- `transport/grpc` stays thin and maps gRPC requests/responses.
- `cmd/main.go` and `internal/app/server.go` act as the composition root and start both HTTP and gRPC servers.

## Bounded Context

`payment-service` owns:

- payment authorization result
- transaction identifier
- payment persistence

`payment-service` does not own:

- order lifecycle
- order cancellation rules
- order persistence

This separation is important for defense: the service stores payment data only and is called by `order-service` through gRPC.

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

## Environment Variables

Default values are listed in [.env.example](/C:/Users/hp/payment-service/.env.example).

- `HTTP_ADDR` default: `:8081`
- `GRPC_ADDR` default: `:50051`
- `POSTGRES_DSN` default: `postgres://postgres@127.0.0.1:55432/payment_service?sslmode=disable`

For local Docker runs, use a DSN with password:

```text
postgres://postgres:postgres@127.0.0.1:55432/payment_service?sslmode=disable
```

## Run

1. Start the payment database:

```bash
docker compose up -d
```

2. Run the service:

```bash
go run ./cmd
```

By default the service starts:

- REST on `:8081`
- gRPC on `:50051`

## REST API Examples

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

## gRPC Contract

`payment-service` implements `PaymentService.ProcessPayment` from the shared protobuf contract stored outside the service repository.

Main request fields:

- `order_id`
- `amount`

Main response fields:

- `payment_id`
- `order_id`
- `status`
- `transaction_id`
- `message`
- `processed_at`

## Architecture Diagram

```mermaid
flowchart LR
    OrderService[Order Service gRPC Client] --> PaymentGRPC[Payment Service gRPC]
    Client --> PaymentHTTP[Payment Service HTTP API]
    PaymentHTTP --> PaymentUC[Payment Use Case]
    PaymentGRPC --> PaymentUC
    PaymentUC --> PaymentRepo[(Payment Service DB)]
```
