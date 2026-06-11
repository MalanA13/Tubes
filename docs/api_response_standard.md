# API Response Standard

> Sprint-02 Task-01 — Standardized API Responses

## 1. Overview

All API endpoints (except `/health`) use a unified JSON response envelope.
This ensures frontend clients can parse every response with a single code path.

## 2. Response Envelope

### Success Response

```json
{
  "success": true,
  "data": { ... }
}
```

| Field     | Type    | Description                        |
| --------- | ------- | ---------------------------------- |
| `success` | boolean | Always `true` for 2xx responses    |
| `data`    | object  | The response payload (varies)      |

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "BAD_REQUEST",
    "message": "invalid request body"
  }
}
```

| Field           | Type    | Description                              |
| --------------- | ------- | ---------------------------------------- |
| `success`       | boolean | Always `false` for 4xx/5xx responses     |
| `error.code`    | string  | Machine-readable error code              |
| `error.message` | string  | Human-readable error description         |

## 3. Error Codes

| HTTP Status | Error Code              |
| ----------- | ----------------------- |
| 400         | `BAD_REQUEST`           |
| 401         | `UNAUTHORIZED`          |
| 403         | `FORBIDDEN`             |
| 404         | `NOT_FOUND`             |
| 405         | `METHOD_NOT_ALLOWED`    |
| 500         | `INTERNAL_SERVER_ERROR` |

## 4. Health Endpoint Exception

`GET /health` is lightweight and does **not** use the standard envelope:

```json
{
  "status": "UP",
  "service": "order"
}
```

## 5. Helper Functions

Package: `internal/response`

| Function              | HTTP Status | Envelope Type |
| --------------------- | ----------- | ------------- |
| `OK(w, data)`         | 200         | Success       |
| `Created(w, data)`    | 201         | Success       |
| `BadRequest(w, msg)`  | 400         | Error         |
| `Unauthorized(w, msg)`| 401         | Error         |
| `Forbidden(w, msg)`   | 403         | Error         |
| `NotFound(w, msg)`    | 404         | Error         |
| `MethodNotAllowed(w, msg)` | 405    | Error         |
| `InternalServerError(w, msg)` | 500 | Error         |

### Usage Example

```go
import "github.com/tubes-cc/logistics/internal/response"

// Success
response.OK(w, map[string]string{"message": "Scan-in berhasil"})

// Error
response.BadRequest(w, "resi_id is required")
```

## 6. Migration Summary

### Files Changed

| File | Change |
| ---- | ------ |
| `internal/response/response.go` | **NEW** — Standard response package |
| `internal/handler/auth_handler.go` | Replaced `http.Error` → `response.*` |
| `internal/handler/order_handler.go` | Replaced `http.Error` → `response.*` |
| `internal/handler/pricing_handler.go` | Replaced `http.Error` → `response.*` |
| `internal/handler/tracking_handler.go` | Replaced `http.Error` + `w.Write` → `response.*` |
| `internal/handler/hub_handler.go` | Replaced `writeError`/`writeSuccess` → `response.*`, removed helpers |
| `internal/handler/courier_handler.go` | Replaced `writeError`/`writeSuccess` → `response.*` |
| `internal/handler/event_handler.go` | Replaced inline JSON error writes → `response.*` |
| `internal/middleware/auth.go` | Replaced `writeJSONError` → `response.*`, removed helper |
| `cmd/order/main.go` | Replaced inline `http.Error` + `http.NotFound` → `response.*` |
| `cmd/hub/main.go` | Replaced `http.Error` in health endpoint → `response.MethodNotAllowed` |
| `cmd/courier/main.go` | Replaced `http.Error` in health endpoint → `response.MethodNotAllowed` |
| `tests/functional/notification_functional_test.go` | Updated to decode new `{success, data}` envelope |
| `docs/openapi.yaml` | Updated all schemas to v2.0.0 with envelope format |
| `docs/api_response_standard.md` | **NEW** — This document |

### Handlers Changed

| Handler | Functions Modified |
| ------- | ------------------ |
| `HandleLogin` | 4 response sites |
| `HandleRegister` | 4 response sites |
| `HandleValidateToken` | 4 response sites |
| `HandleOrderHTTP` | 3 response sites |
| `HandleSendPricingHTTP` | 3 response sites |
| `HandleSendTrackingHTTP` | 3 response sites |
| `HandleSendNotificationHTTP` | 4 response sites |
| `HubHandler.ScanIn` | 4 response sites |
| `HubHandler.ScanOut` | 4 response sites |
| `CourierHandler.AssignCourier` | 4 response sites |
| `CourierHandler.UpdateDeliveryStatus` | 4 response sites |
| `AuthMiddleware.Authenticate` | 3 response sites |
| `AuthMiddleware.RequireRole` | 2 response sites |

**Total: 13 handlers, 44 response sites migrated.**

### Removed Code

- `writeSuccess()` — was in `hub_handler.go`
- `writeError()` — was in `hub_handler.go`
- `writeJSONError()` — was in `middleware/auth.go`
- `successResponse` struct — was in `hub_handler.go`
- `errorResponse` struct — was in `hub_handler.go`
- `middlewareErrorResponse` struct — was in `middleware/auth.go`

## 7. Backward Compatibility Impact

> **Breaking Change** for existing API consumers.

| Before | After |
| ------ | ----- |
| `{"token": "..."}` | `{"success": true, "data": {"token": "..."}}` |
| `{"order_id": "..."}` | `{"success": true, "data": {"order_id": "..."}}` |
| `plain text error` | `{"success": false, "error": {"code": "...", "message": "..."}}` |
| `{"error": "Bad Request", "message": "..."}` | `{"success": false, "error": {"code": "BAD_REQUEST", "message": "..."}}` |

Frontend clients must update their JSON parsing to unwrap the `data` field for success responses and check the `success` boolean before accessing `data` or `error`.

## 8. Test Results

```
go test ./... → PASS (all packages)
go build ./cmd/auth → OK
go build ./cmd/order → OK
go build ./cmd/pricing → OK
go build ./cmd/tracking → OK
go build ./cmd/hub → OK
go build ./cmd/courier → OK
go build ./cmd/notification → OK
```
