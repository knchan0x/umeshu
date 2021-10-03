package umeshu

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/knchan0x/umeshu/log"
)

// Recovery is a middleware to recover umeshu engine
// from panic error and provides log for tracing.
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Recovery("server has been auto recovered from error: %s\n\n", trace(fmt.Sprintf("%s", err)))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()

		c.Next()
	}
}

// trace provides traceback message.
func trace(msg string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])

	var str strings.Builder
	str.WriteString(msg + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
