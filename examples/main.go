package main

import (
	"math/rand"
	"time"

	"github.com/LokiWager/analysis-demo/core"
	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/go-resty/resty/v2"
)

// This is a simple example to show how to use the analysis tool.
func main() {
	diagnostics := core.NewDiagnostic(&core.DiagnosticConfig{})
	go func() {
		diagnostics.Start()
	}()

	// Do something else
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	ticker2 := time.NewTicker(60 * time.Second)
	defer ticker2.Stop()

	for {
		select {
		case <-ticker.C:
			// Do something
			go func() {
				time.Sleep(5 * time.Second)
				logger.Infof("sleep 5 seconds")
			}()

			n := rand.Intn(1000)
			sum := 0
			for i := 0; i < n; i++ {
				sum += i
			}
			logger.Infof("sum is %d", sum)
		case <-ticker2.C:
			restyClient := resty.New()
			// call google.com
			_, err := restyClient.R().Get("https://www.google.com")
			if err != nil {
				logger.Errorf("call google.com failed: %v", err)
			} else {
				logger.Infof("call google.com success")
			}
		}
	}
}
