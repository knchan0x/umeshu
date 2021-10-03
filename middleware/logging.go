package middleware

import (
	"time"

	"github.com/knchan0x/umeshu"
	"github.com/knchan0x/umeshu/log"
)

// Logging logs the time used for responsing a http request
func Logging() umeshu.HandlerFunc {
	return func(c *umeshu.Context) {
		t := time.Now()
		c.Next()
		defer log.Info("%v | %d | %s %s", time.Since(t), c.StatusCode, c.Method, c.Path)
	}
}
