package middleware

import (
	"crypto/subtle"

	"google.golang.org/grpc"
)

//go:generate go-enum -f=$GOFILE --nocase --values

type Middleware interface {
	GetInterceptor() grpc.UnaryServerInterceptor
	GetType() Type
}

// Type
// ENUM(
// none
// basic
// )
type Type string

func cryptCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ContextKey
// ENUM(
// authorization
// username
// )
type ContextKey string
