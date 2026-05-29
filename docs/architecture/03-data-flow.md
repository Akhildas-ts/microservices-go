# Request Data Flow

> Walks two real requests end-to-end so you can trace any future request the same way.

## Example 1 — Customer login

```
1. POST /login (email,password) ────────────────────────┐
                                                        │
                                  ┌─────────────────────▼─────────┐
                                  │ API Gateway (Gin)             │
                                  │ pkg/auth/routes.go: Login()   │
                                  └────────────┬──────────────────┘
                                               │ gRPC Login(LoginRequest)
                                               ▼
                                  ┌────────────────────────────────┐
                                  │ Auth Service (gRPC)            │
                                  │ pkg/services/auth.go: Login()  │
                                  └────────────┬───────────────────┘
                                               │ SELECT * FROM users WHERE email=?
                                               ▼
                                          ┌─────────┐
                                          │ Postgres│
                                          │ e_auth_ │
                                          │   svc   │
                                          └─────────┘
                                               │ row
                                               ▼
                                  ┌────────────────────────────────┐
                                  │ Auth Service                   │
                                  │ - bcrypt compare               │
                                  │ - JWT sign (HS256)             │
                                  │ - return LoginResponse{token}  │
                                  └────────────┬───────────────────┘
                                               ▼
                                  ┌────────────────────────────────┐
                                  │ API Gateway                    │
                                  │ - wrap response in JSON        │
                                  │ - return 200 + token to client │
                                  └────────────────────────────────┘
```

## Example 2 — Add item to cart (the multi-hop case)

```
1. POST /cart/items (productId, qty)  Authorization: Bearer <jwt>
                                                          │
                                  ┌───────────────────────▼───────┐
                                  │ API Gateway                   │
                                  │ - middleware: validate JWT    │
                                  │   (calls auth-svc.Validate)   │
                                  └─────────────┬─────────────────┘
                                                │ gRPC AddItem(...)
                                                ▼
                                  ┌────────────────────────────────┐
                                  │ Cart Service                   │
                                  │ pkg/services/cart.go           │
                                  │ - calls Product-Svc to check   │
                                  │   the product exists / price   │
                                  └────────┬────────────┬──────────┘
                                           │ gRPC       │ SQL
                                           ▼            ▼
                                  ┌────────────────┐ ┌──────────┐
                                  │ Product Service│ │ Postgres │
                                  │ FindOne()      │ │ e_cart_  │
                                  └───────┬────────┘ │   svc    │
                                          │          └──────────┘
                                          ▼ row from e_product_svc
                                  ┌────────────────┐
                                  │ Postgres       │
                                  │ e_product_svc  │
                                  └────────────────┘

   Cart-svc returns cart state → Gateway → Client (JSON)
```

## Why this matters

This shape — gateway as a fan-out node, services as **mostly-leaf** processes that call at most one other service — is what keeps the system understandable. The moment a service starts calling 3+ others synchronously, you have a distributed monolith. See [`microservices/03-boundaries.md`](../microservices/03-boundaries.md).

## How to trace a NEW endpoint
1. Find the route in `go-grpc-api-gateway/pkg/<domain>/routes.go`.
2. Note the handler — it calls a gRPC client method.
3. Follow that method into the target service's `pkg/services/*.go`.
4. From there, follow GORM calls into the DB models.
5. If the service calls another service, the gRPC client lives in `pkg/client/`.
