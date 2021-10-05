package umeshu

import (
	"time"

	"github.com/knchan0x/umeshu/log"
)

// Logging logs the time used for responsing a http request
func Logging() HandlerFunc {
	return func(c *Context) {
		t := time.Now()
		c.Next()
		defer log.Info("[Umeshu] %v | %d | %s %s", time.Since(t), c.StatusCode, c.Method, c.Path)
	}
}
