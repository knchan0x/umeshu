package umeshu

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/knchan0x/umeshu/log"
)

func BenchmarkNewContext_Pool(b *testing.B) {
	app := Default()
	app.GET("/", func(c *Context) {
		c.String(200, "Hi")
	})

	go func() {
		time.Sleep(60 * time.Second)
		app.Shutdown()
	}()

	myclient := &http.Client{}
	go func() {
		time.Sleep(5 * time.Second)
		for i := 1; i <= 50; i++ {
			go func(i int) {
				url := fmt.Sprintf("http://localhost:8080/")
				resp, err := myclient.Get(url)
				if err != nil {
					log.Error("client get error: %s", err)
					return
				}
				if resp.StatusCode == http.StatusOK {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Error("read body error: %s", err.Error())
					}
					bodyString := string(bodyBytes)
					if bodyString != "Hi" {
						log.Error("error body != hi, body = %s", bodyString)
					}
				}
				if errBody := resp.Body.Close(); errBody != nil {
					log.Error("error on closing body: %s", errBody.Error())
				}
			}(i)
			if i == 50 {
				log.Info("%d times done", i)
			}
		}
	}()
	app.Run(":8080")
}
