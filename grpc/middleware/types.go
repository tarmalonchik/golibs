package middleware

//go:generate go-enum -f=$GOFILE --nocase --values

// MiddlewareType
// ENUM(
// none
// basic
// )
type MiddlewareType string
