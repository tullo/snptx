package main

// the key must be unexported type to avoid collisions
type contextKey string

const isAuthenticatedContextKey = contextKey("isAuthenticated")
