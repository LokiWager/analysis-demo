package main

import (
	"net/http"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/LokiWager/analysis-demo/pkg/logger"
	"github.com/LokiWager/analysis-demo/pkg/rest"
	"github.com/LokiWager/analysis-demo/pkg/service"

	_ "net/http/pprof"
)

func main() {
	logger.Init(&logger.Config{Debug: false})
	app := &cli.App{
		Name:  "diagnostic",
		Usage: "diagnostic is a tool to diagnose the running process",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:  "pid",
				Value: 0,
				Usage: "The process id to diagnose",
			},
			&cli.BoolFlag{
				Name:  "persist",
				Value: false,
				Usage: "The path to persist the diagnostic data",
			},
		},
		Action: func(c *cli.Context) error {
			pid := c.Int("pid")
			if pid == 0 {
				pid = os.Getpid()
			}
			config := &service.ServiceConfig{
				ProcessID: pid,
			}

			if c.Bool("persist") {
				config.Persist = true
			}

			rest.New(config).ServerForever(8080)
			return nil
		},
	}

	go func() {
		err := http.ListenAndServe("localhost:6060", nil)
		if err != nil {
			logger.Fatalf("start pprof failed: %v", err)
		}
	}()

	logger.Info("diagnostic server starting")
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
