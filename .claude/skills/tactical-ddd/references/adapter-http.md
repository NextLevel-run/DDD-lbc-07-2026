# HTTP Driving Adapters

## Concept

**HTTP Handlers** are **driving adapters** that:

- Receive HTTP requests and translate them to command/query calls
- Map domain errors to appropriate HTTP status codes
- Transform view models to JSON responses
- Handle request validation and parsing

HTTP handlers are thin translation layers - no business logic.

## Implementation Pattern

### File Structure
```
internal/{bounded-context}/adapter/driving/http/
├── handler.go     # Handler struct and methods
├── dto.go         # Request/Response DTOs
└── handler_test.go
```

### DTOs (Data Transfer Objects)

```go
// internal/order/adapter/driving/http/dto.go

package http

import "time"

// Request DTOs

type OrderItemRequest struct {
    ProductID string `json:"productId"`
    Quantity  int    `json:"quantity"`
    UnitPrice int64  `json:"unitPrice"`
    Currency  string `json:"currency"`
}

type PlaceOrderRequest struct {
    CustomerEmail string             `json:"customerEmail"`
    Street        string             `json:"street"`
    City          string             `json:"city"`
    PostalCode    string             `json:"postalCode"`
    Country       string             `json:"country"`
    Items         []OrderItemRequest `json:"items"`
}

type CancelOrderRequest struct {
    CancellationReason string `json:"cancellationReason"`
}

type NotifyCustomerRequest struct {
    Subject string `json:"subject"`
    Message string `json:"message"`
}

// Response DTOs

type PlaceOrderResponse struct {
    ID      string `json:"id"`
    Message string `json:"message"`
}

type OrderViewResponse struct {
    Id              string                  `json:"id"`
    Status          string                  `json:"status"`
    ShippingAddress AddressResponse         `json:"shippingAddress"`
    Items           []OrderItemViewResponse `json:"items"`
    TotalAmount     int64                   `json:"totalAmount"`
    TotalCurrency   string                  `json:"totalCurrency"`
    PlacedAt        time.Time               `json:"placedAt"`
}

type AddressResponse struct {
    Street     string `json:"street"`
    City       string `json:"city"`
    PostalCode string `json:"postalCode"`
    Country    string `json:"country"`
}

type OrderItemViewResponse struct {
    ProductID string `json:"productId"`
    Quantity  int    `json:"quantity"`
    UnitPrice int64  `json:"unitPrice"`
    Currency  string `json:"currency"`
}

type OrderListItemResponse struct {
    Id            string    `json:"id"`
    Status        string    `json:"status"`
    City          string    `json:"city"`
    TotalAmount   int64     `json:"totalAmount"`
    TotalCurrency string    `json:"totalCurrency"`
    PlacedAt      time.Time `json:"placedAt"`
}

type OrdersListResponse struct {
    Orders []OrderListItemResponse `json:"orders"`
}

type ErrorResponse struct {
    Error string `json:"error"`
}

type CancelOrderResponse struct {
    Message string `json:"message"`
}

type NotifyCustomerResponse struct {
    Message string `json:"message"`
}
```

### Handler Structure

```go
// internal/order/adapter/driving/http/handler.go

package http

import (
    "ddd-second-hand-marketplace/internal/order/application/command"
    "ddd-second-hand-marketplace/internal/order/application/query"
    "ddd-second-hand-marketplace/internal/order/domain"
    "encoding/json"
    "errors"
    "net/http"
)

// Handler holds all command and query dependencies
type Handler struct {
    placeOrderCommand      command.PlaceOrderCommand
    cancelOrderCommand     command.CancelOrderCommand
    notifyCustomerCommand  command.NotifyCustomerCommand
    getOrderQuery          query.GetOrderQuery
    findOrdersQuery        query.FindOrdersQuery
}

// NewHandler creates a handler with all dependencies injected
func NewHandler(
    placeOrderCommand command.PlaceOrderCommand,
    cancelOrderCommand command.CancelOrderCommand,
    notifyCustomerCommand command.NotifyCustomerCommand,
    getOrderQuery query.GetOrderQuery,
    findOrdersQuery query.FindOrdersQuery,
) *Handler {
    return &Handler{
        placeOrderCommand:      placeOrderCommand,
        cancelOrderCommand:     cancelOrderCommand,
        notifyCustomerCommand:  notifyCustomerCommand,
        getOrderQuery:          getOrderQuery,
        findOrdersQuery:        findOrdersQuery,
    }
}
```

### Command Handler (POST)

```go
// PlaceOrder handles POST /orders
func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
    // 1. Validate HTTP method
    if r.Method != http.MethodPost {
        h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // 2. Decode request body
    var req PlaceOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
        return
    }

    // 3. Map DTO to command args
    items := make([]command.OrderItemArgs, len(req.Items))
    for i, item := range req.Items {
        items[i] = command.OrderItemArgs{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            UnitPrice: item.UnitPrice,
            Currency:  item.Currency,
        }
    }

    // 4. Execute command
    orderID, err := h.placeOrderCommand(command.PlaceOrderCommandArgs{
        CustomerEmail: req.CustomerEmail,
        Street:        req.Street,
        City:          req.City,
        PostalCode:    req.PostalCode,
        Country:       req.Country,
        Items:         items,
    })

    // 5. Map domain errors to HTTP status codes
    if err != nil {
        h.handleDomainError(w, err)
        return
    }

    // 6. Return success response
    h.respondSuccess(w, PlaceOrderResponse{
        ID:      orderID,
        Message: "Order placed successfully",
    }, http.StatusCreated)
}
```

### Query Handler (GET Single)

```go
// GetOrder handles GET /orders?id={id}
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Extract ID from query parameter
    id := r.URL.Query().Get("id")
    if id == "" {
        h.respondError(w, "Missing 'id' query parameter", http.StatusBadRequest)
        return
    }

    // Execute query
    orderView, err := h.getOrderQuery(id)
    if err != nil {
        h.respondError(w, "Order not found", http.StatusNotFound)
        return
    }

    // Map view model to response DTO
    items := make([]OrderItemViewResponse, len(orderView.Items))
    for i, item := range orderView.Items {
        items[i] = OrderItemViewResponse{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
            UnitPrice: item.UnitPrice,
            Currency:  item.Currency,
        }
    }

    response := OrderViewResponse{
        Id:     orderView.Id,
        Status: orderView.Status,
        ShippingAddress: AddressResponse{
            Street:     orderView.ShippingAddress.Street,
            City:       orderView.ShippingAddress.City,
            PostalCode: orderView.ShippingAddress.PostalCode,
            Country:    orderView.ShippingAddress.Country,
        },
        Items:         items,
        TotalAmount:   orderView.TotalAmount,
        TotalCurrency: orderView.TotalCurrency,
        PlacedAt:      orderView.PlacedAt,
    }

    h.respondSuccess(w, response, http.StatusOK)
}
```

### Query Handler (GET List with Filters)

```go
// FindOrders handles GET /orders with optional filters
func (h *Handler) FindOrders(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        h.respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Parse optional query parameters
    queryParams := r.URL.Query()

    var status *string
    if s := queryParams.Get("status"); s != "" {
        status = &s
    }

    var minTotal *int64
    if minTotalStr := queryParams.Get("minTotal"); minTotalStr != "" {
        val, err := strconv.ParseInt(minTotalStr, 10, 64)
        if err != nil {
            h.respondError(w, "Invalid 'minTotal' parameter", http.StatusBadRequest)
            return
        }
        minTotal = &val
    }

    // Execute query
    listItems, err := h.findOrdersQuery(query.FindOrdersArgs{
        Status:    status,
        MinTotal:  minTotal,
        SortBy:    queryParams.Get("sortBy"),
        SortOrder: queryParams.Get("sortOrder"),
    })

    if err != nil {
        h.respondError(w, "Error retrieving orders: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Map to response DTOs
    responseItems := make([]OrderListItemResponse, len(listItems))
    for i, item := range listItems {
        responseItems[i] = OrderListItemResponse{
            Id:            item.Id,
            Status:        item.Status,
            City:          item.City,
            TotalAmount:   item.TotalAmount,
            TotalCurrency: item.TotalCurrency,
            PlacedAt:      item.PlacedAt,
        }
    }

    h.respondSuccess(w, OrdersListResponse{Orders: responseItems}, http.StatusOK)
}
```

### Error Mapping

```go
// handleDomainError maps domain errors to HTTP responses
func (h *Handler) handleDomainError(w http.ResponseWriter, err error) {
    statusCode := http.StatusInternalServerError
    errorMessage := err.Error()

    switch {
    case errors.Is(err, domain.ErrInvalidEmail):
        statusCode = http.StatusBadRequest
        errorMessage = "Invalid email format"
    case errors.Is(err, domain.ErrEmptyOrderItems):
        statusCode = http.StatusBadRequest
        errorMessage = "Order must have at least one item"
    case errors.Is(err, domain.ErrInvalidQuantity):
        statusCode = http.StatusBadRequest
        errorMessage = "Item quantity must be greater than zero"
    case errors.Is(err, domain.ErrOrderNotFound):
        statusCode = http.StatusNotFound
        errorMessage = "Order not found"
    case errors.Is(err, domain.ErrOrderAlreadyShipped):
        statusCode = http.StatusConflict
        errorMessage = "Cannot modify shipped order"
    }

    h.respondError(w, errorMessage, statusCode)
}
```

### Response Helpers

```go
func (h *Handler) respondError(w http.ResponseWriter, message string, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (h *Handler) respondSuccess(w http.ResponseWriter, data interface{}, statusCode int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(data)
}
```

## Key Principles

1. **Thin handlers** - just translation, no business logic
2. **DTOs separate from domain** - decouple API contract from domain
3. **Error mapping** - translate domain errors to HTTP status codes
4. **Constructor injection** - all commands/queries injected via `NewHandler`
5. **Standard library** - no external HTTP framework required

## Route Registration

```go
// cmd/api/main.go

func main() {
    // ... create commands and queries ...

    handler := http.NewHandler(
        placeOrderCommand,
        cancelOrderCommand,
        notifyCustomerCommand,
        getOrderQuery,
        findOrdersQuery,
    )

    // Register routes
    mux := http.NewServeMux()
    mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "POST":
            handler.PlaceOrder(w, r)
        case "GET":
            if r.URL.Query().Get("id") != "" {
                handler.GetOrder(w, r)
            } else {
                handler.FindOrders(w, r)
            }
        case "DELETE":
            handler.CancelOrder(w, r)
        }
    })

    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

## Testing

### Test Setup

```go
// internal/order/adapter/driving/http/handler_test.go

package http

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// Mock command for testing
type mockPlaceOrderCommand struct {
    returnId  string
    returnErr error
    calledWith command.PlaceOrderCommandArgs
}

func (m *mockPlaceOrderCommand) execute(args command.PlaceOrderCommandArgs) (string, error) {
    m.calledWith = args
    return m.returnId, m.returnErr
}

func setupHandlerTest(t *testing.T) (*Handler, *mockPlaceOrderCommand) {
    t.Helper()

    mockCmd := &mockPlaceOrderCommand{returnId: "test-id-123"}

    handler := &Handler{
        placeOrderCommand: mockCmd.execute,
    }

    return handler, mockCmd
}
```

### Success Test

```go
func TestPlaceOrder_Success(t *testing.T) {
    // Given
    handler, mockCmd := setupHandlerTest(t)
    mockCmd.returnId = "generated-id-123"

    body := PlaceOrderRequest{
        CustomerEmail: "customer@example.com",
        Street:        "123 Main St",
        City:          "Paris",
        PostalCode:    "75001",
        Country:       "France",
        Items: []OrderItemRequest{
            {ProductID: "PROD-1", Quantity: 2, UnitPrice: 2999, Currency: "EUR"},
        },
    }
    jsonBody, _ := json.Marshal(body)

    req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    // When
    handler.PlaceOrder(rec, req)

    // Then
    assert.Equal(t, http.StatusCreated, rec.Code)

    var response PlaceOrderResponse
    err := json.NewDecoder(rec.Body).Decode(&response)
    require.NoError(t, err)

    assert.Equal(t, "generated-id-123", response.ID)
    assert.Equal(t, "Order placed successfully", response.Message)

    // Verify command was called with correct args
    assert.Equal(t, "customer@example.com", mockCmd.calledWith.CustomerEmail)
    assert.Equal(t, "Paris", mockCmd.calledWith.City)
}
```

### Error Test

```go
func TestPlaceOrder_ValidationError(t *testing.T) {
    // Given
    handler, mockCmd := setupHandlerTest(t)
    mockCmd.returnErr = domain.ErrEmptyOrderItems

    body := PlaceOrderRequest{
        CustomerEmail: "customer@example.com",
        Street: "123 Main St", City: "Paris", PostalCode: "75001", Country: "France",
        Items: []OrderItemRequest{},
    }
    jsonBody, _ := json.Marshal(body)

    req := httptest.NewRequest("POST", "/orders", bytes.NewBuffer(jsonBody))
    rec := httptest.NewRecorder()

    // When
    handler.PlaceOrder(rec, req)

    // Then
    assert.Equal(t, http.StatusBadRequest, rec.Code)

    var response ErrorResponse
    _ = json.NewDecoder(rec.Body).Decode(&response)
    assert.Equal(t, "Order must have at least one item", response.Error)
}
```

### Method Not Allowed Test

```go
func TestPlaceOrder_MethodNotAllowed(t *testing.T) {
    // Given
    handler, _ := setupHandlerTest(t)

    req := httptest.NewRequest("GET", "/orders", nil)
    rec := httptest.NewRecorder()

    // When
    handler.PlaceOrder(rec, req)

    // Then
    assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}
```

### Invalid JSON Test

```go
func TestPlaceOrder_InvalidJSON(t *testing.T) {
    // Given
    handler, _ := setupHandlerTest(t)

    req := httptest.NewRequest("POST", "/orders", bytes.NewBufferString("not json"))
    rec := httptest.NewRecorder()

    // When
    handler.PlaceOrder(rec, req)

    // Then
    assert.Equal(t, http.StatusBadRequest, rec.Code)

    var response ErrorResponse
    _ = json.NewDecoder(rec.Body).Decode(&response)
    assert.Contains(t, response.Error, "Invalid request body")
}
```

## Testing Principles

1. **Use httptest** - `httptest.NewRequest`, `httptest.NewRecorder`
2. **Mock commands/queries** - don't test domain logic in HTTP tests
3. **Test HTTP concerns** - status codes, headers, JSON encoding
4. **Test error mapping** - domain errors → HTTP status codes
5. **Test edge cases** - wrong method, invalid JSON, missing params
