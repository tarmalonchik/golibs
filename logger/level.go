//go:generate go-enum -f=$GOFILE --nocase --values
package logger

// Level
// ENUM(
// panic
// fatal
// error
// warn
// info
// debug
// )
type Level string