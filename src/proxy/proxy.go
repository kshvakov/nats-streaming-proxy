package proxy

import (
	"bufio"
	"net"
	"os"
	"os/signal"
	"syscall"

	nats "github.com/nats-io/go-nats-streaming"
	log "github.com/sirupsen/logrus"
)

const version = "1.0.1"

type Options struct {
	ServerAddr    string
	MetricsAddr   string
	NatsClusterID string
	NatsClientID  string
	NatsURL       string
}

// New NATS Steaming proxy
func New(options Options) (_ *Proxy, err error) {
	var (
		proxy = Proxy{
			address: options.ServerAddr,
			signals: make(chan os.Signal),
		}
		connOpts = []nats.Option{
			nats.Pings(10, 120),
			nats.NatsURL(options.NatsURL),
			nats.SetConnectionLostHandler(func(conn nats.Conn, err error) {
				log.Errorf("lost connection: %v", err)
				proxy.signals <- syscall.SIGTERM
			}),
		}
	)
	if proxy.natsConn, err = nats.Connect(options.NatsClusterID, options.NatsClientID, connOpts...); err != nil {
		return nil, err
	}
	log.Infof("listen=%s, nats-cluster-id=%s, nats-client-id=%s", options.ServerAddr, options.NatsClusterID, options.NatsClientID)
	go metrics(options.MetricsAddr)
	go proxy.waitSignal()
	return &proxy, nil
}

type Proxy struct {
	address  string
	signals  chan os.Signal
	natsConn nats.Conn
}

func (p *Proxy) waitSignal() {
	signal.Notify(p.signals,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	select {
	case sig := <-p.signals:
		log.Infof("shutdown [%s]", sig)
		p.natsConn.Close()
		{
			os.Exit(0)
		}
	}
}

// Listen announces on the local network address.
func (p *Proxy) Listen() error {
	listener, err := net.Listen("tcp", p.address)
	if err != nil {
		return err
	}
	for {
		if conn, err := listener.Accept(); err == nil {
			go (&connect{
				net:      conn,
				natsConn: p.natsConn,
				buffer: bufio.NewReadWriter(
					bufio.NewReaderSize(conn, 8*1024),
					bufio.NewWriter(conn),
				),
			}).serve()
		}
	}
}
