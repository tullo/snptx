package web

import "time"

// ctxKey represents the type of value for the context key.
type ctxKey uint

// KeyValues is how request values or stored/retrieved.
const KeyValues ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}
