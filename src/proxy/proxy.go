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

type Options struct {
	ServerAddr       string
	MetricsAddr      string
	NatsClusterID    string
	NatsClientID     string
	NatsPublishAsync bool
	NatsURL          string
}

// New NATS Steaming proxy
func New(version string, options Options) (_ *Proxy, err error) {
	var (
		proxy = Proxy{
			version: version,
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
	if proxy.nats.conn, err = nats.Connect(options.NatsClusterID, options.NatsClientID, connOpts...); err != nil {
		return nil, err
	}
	proxy.nats.async = options.NatsPublishAsync
	log.Infof("listen=%s, nats-cluster-id=%s, nats-client-id=%s, nats-publish-async=%t",
		options.ServerAddr,
		options.NatsClusterID,
		options.NatsClientID,
		proxy.nats.async,
	)
	go metrics(options.MetricsAddr)
	go proxy.waitSignal()
	return &proxy, nil
}

type Proxy struct {
	version string
	address string
	signals chan os.Signal
	nats    struct {
		conn  nats.Conn
		async bool
	}
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
		p.nats.conn.Close()
		{
			os.Exit(0)
		}
	}
}

func (p *Proxy) publish(subject string, data []byte) (err error) {
	switch {
	case p.nats.async:
		_, err = p.nats.conn.PublishAsync(subject, data, nil)
	default:
		err = p.nats.conn.Publish(subject, data)
	}
	return err
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
				net:     conn,
				version: p.version,
				publish: p.publish,
				buffer: bufio.NewReadWriter(
					bufio.NewReaderSize(conn, 8*1024),
					bufio.NewWriter(conn),
				),
			}).serve()
		}
	}
}
