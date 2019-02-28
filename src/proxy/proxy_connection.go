package proxy

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	nats "github.com/nats-io/go-nats-streaming"
	log "github.com/sirupsen/logrus"
)

var (
	crlf      = []byte("\r\n")
	noreply   = []byte("noreply")
	startTime = time.Now()
)

var (
	StatusEnd       = []byte("END\r\n")
	StatusStored    = []byte("STORED\r\n")
	StatusNotStored = []byte("NOT_STORED\r\n")
	StatPattern     = "STAT %s %v\r\n"
)

type connect struct {
	net      net.Conn
	version  string
	buffer   *bufio.ReadWriter
	publish  func(string, []byte) error
	natsConn nats.Conn
}

func (conn *connect) serve() {
	address := conn.net.RemoteAddr().String()
	{
		connectionsInc(address)
	}
	defer func() {
		connectionsDec(address)
		conn.Close()
	}()
	for {
		switch err := conn.handle(); {
		case err != nil:
			if err != io.EOF {
				log.Errorf("handle %v", err)
			}
			return
		default:
			if err := conn.buffer.Flush(); err != nil {
				return
			}
		}
	}
}

func (conn *connect) handle() error {
	line, _, err := conn.buffer.ReadLine()
	if err != nil || len(line) < 4 {
		return io.EOF
	}
	switch line[0] {
	case 'g': // get
		keys := strings.Fields(string(line[4:]))
		if len(keys) == 0 {
			return io.EOF
		}
		for _, key := range keys {
			conn.buffer.Write([]byte("VALUE " + key + " 0 13\r\n"))
			conn.buffer.Write([]byte("not supported"))
			conn.buffer.Write(crlf)
		}
		conn.buffer.Write(StatusEnd)
	case 'q': // quit
		return io.EOF
	case 's':
		switch line[1] {
		case 'e': // set
			var (
				fields    = bytes.Fields(line[4:])
				subject   = string(fields[0])
				length, _ = strconv.Atoi(string(fields[3]))
				value     = make([]byte, length+2)
				n, err    = conn.buffer.Read(value)
			)
			switch {
			case err != nil && err != io.EOF:
				return err
			case !bytes.HasSuffix(value, crlf):
				return io.EOF
			case n != length+2: // bad chunk
				return io.EOF
			}
			switch err := conn.publish(subject, value[:length]); {
			case err != nil:
				conn.net.Write(StatusNotStored)
				{
					reqFailedProm.WithLabelValues(subject).Inc()
				}
			default:
				conn.net.Write(StatusStored)
				{
					reqProcessedInc(subject)
				}
			}
		case 't': // stats
			fmt.Fprintf(conn.buffer, StatPattern, "pid", os.Getpid())
			fmt.Fprintf(conn.buffer, StatPattern, "time", time.Now().Unix())
			fmt.Fprintf(conn.buffer, StatPattern, "server", "nats-streaming-proxy")
			fmt.Fprintf(conn.buffer, StatPattern, "uptime", int64(time.Since(startTime).Seconds()))
			fmt.Fprintf(conn.buffer, StatPattern, "version", conn.version)
			fmt.Fprintf(conn.buffer, StatPattern, "num_goroutine", runtime.NumGoroutine())
			fmt.Fprintf(conn.buffer, StatPattern, "cmd_set", atomic.LoadInt64(&reqProcessed))
			fmt.Fprintf(conn.buffer, StatPattern, "curr_connections", atomic.LoadInt64(&currentConnections))
			fmt.Fprintf(conn.buffer, StatPattern, "total_connections", atomic.LoadInt64(&totalConnections))
			{
				conn.buffer.Write(StatusEnd)
			}
		}
	case 'v': // version
		conn.net.Write([]byte("VERSION " + conn.version + "\r\n"))
	default:
		return io.EOF
	}
	return nil
}

func (conn *connect) Close() error {
	conn.buffer.Flush()
	return conn.net.Close()
}
