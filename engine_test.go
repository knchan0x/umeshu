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
		for i := 0; i < 5000; i++ {
			go func(i int) {
				url := fmt.Sprintf("http://localhost:8080/")
				resp, _ := myclient.Get(url)
				defer resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					bodyBytes, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Error(err.Error())
					}
					bodyString := string(bodyBytes)
					log.SetLevel(log.InfoLevel)
					log.Debug(bodyString)
				}
			}(i)
			if i == 4999 {
				log.Info("%d times done", i)
			}
		}
	}()
	app.Run(":8080")
}
