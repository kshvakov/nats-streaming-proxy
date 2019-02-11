package main

import (
	"fmt"
	"os"

	"github.com/kshvakov/nats-streaming-proxy/src/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	BuildDate            string
	GitBranch, GitCommit string
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
}
func main() {
	app := cli.NewApp()
	app.Name = "NATS Streaming memcache proxy"
	app.Usage = "."
	app.Version = fmt.Sprintf("rev[%s] %s (%s UTC).", GitCommit, GitBranch, BuildDate)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "nats-url",
			EnvVar: "NATS_URL",
			Value:  "nats://127.0.0.1:4222",
			Usage:  "The url can contain username/password semantics. e.g. nats://username:pass@localhost:4222.",
		},
		cli.StringFlag{
			Name:   "nats-client-id",
			EnvVar: "NATS_CLIENT_ID",
			Value:  "nats-streaming-proxy",
			Usage:  "Unique client ID. ClientID can contain only alphanumeric and `-` or `_` characters.",
		},
		cli.StringFlag{
			Name:   "nats-cluster-id",
			EnvVar: "NATS_CLUSTER_ID",
			Value:  "test-cluster",
			Usage:  "ID of the NATS Streaming cluster.",
		},
		cli.StringFlag{
			Name:   "log-level",
			EnvVar: "LOG_LEVEL",
			Value:  "info",
			Usage:  "",
		},
		cli.StringFlag{
			Name:   "server-addr",
			EnvVar: "SERVER_ADDR",
			Value:  ":11211",
			Usage:  "Listen address.",
		},
		cli.StringFlag{
			Name:   "metrics-addr",
			EnvVar: "METRICS_ADDR",
			Value:  ":1414",
			Usage:  "Prometheus metrics HTTP endpoint.",
		},
		cli.BoolFlag{
			Name:   "debug",
			EnvVar: "DEBUG",
			Usage:  "Enable debug mode.",
		},
	}
	app.Action = func(c *cli.Context) error {
		switch {
		case c.Bool("debug"):
			log.SetLevel(log.DebugLevel)
		default:
			if level, err := log.ParseLevel(c.String("log-level")); err == nil {
				log.SetLevel(level)
			}
		}
		proxy, err := proxy.New(proxy.Options{
			NatsURL:       c.String("nats-url"),
			NatsClientID:  c.String("nats-client-id"),
			NatsClusterID: c.String("nats-cluster-id"),
			NatsGroup:     c.String("nats-group"),
			MetricsAddr:   c.String("metrics-addr"),
			ServerAddr:    c.String("server-addr"),
		})
		if err != nil {
			return err
		}
		return proxy.Listen()
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
