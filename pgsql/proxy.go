package pgsql

import (
	"github.com/orderbynull/protocol/pgsql"
	"github.com/rs/xid"
	"io"
	"log"
	"net"
)

// Proxy ...
type Proxy struct {
	listen   string
	target   string
	messages chan interface{}
}

// NewProxy creates new instance of Proxy
func NewProxy(listen string, target string) *Proxy {
	return &Proxy{listen: listen, target: target}
}

// Run runs proxy server on specified port and handles each incoming
// tcp connection in separate goroutine.
func (p *Proxy) Run() chan interface{} {
	p.messages = make(chan interface{})

	go func() {
		listener, err := net.Listen("tcp", p.listen)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := listener.Close(); err != nil {
				log.Println(err)
			}
		}()

		for {
			client, err := listener.Accept()
			if err != nil {
				log.Print(err.Error())
			}

			go p.handleConnection(client)
		}
	}()

	return p.messages
}

// handleConnection makes connection to target host per each incoming tcp connection
// and forwards all traffic from source to target.
func (p *Proxy) handleConnection(in io.ReadWriteCloser) {
	defer func() {
		if err := in.Close(); err != nil {
			log.Println(err)
		}
	}()

	out, err := net.Dial("tcp", p.target)
	if err != nil {
		log.Print(err)
		return
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Println(err)
		}
	}()

	err = p.proxyTraffic(in, out)
	if err != nil {
		log.Println(err)
	}
}

// proxyTraffic ...
func (p *Proxy) proxyTraffic(client, server io.ReadWriteCloser) error {
	connId := xid.New().String()

	requestWatcher := &watcher{p, connId, true, pgsql.PacketBuilder{}}
	responseWatcher := &watcher{p, connId, false, pgsql.PacketBuilder{}}

	// Copy bytes from client to server
	go func() {
		if _, err := io.Copy(io.MultiWriter(server, requestWatcher), client); err != nil {
			log.Println(err)
		}
	}()

	// Copy bytes from server to client
	if _, err := io.Copy(io.MultiWriter(client, responseWatcher), server); err != nil {
		log.Println(err)
	}

	return nil
}

// watcher ...
type watcher struct {
	proxy     *Proxy
	connId    string
	isRequest bool
	builder   pgsql.PacketBuilder
}

// Write ...
func (rrw *watcher) Write(p []byte) (n int, err error) {
	packet, err := rrw.builder.Build(p)
	if err != nil {
		println(err)
	}
	if packet != nil {
		for _, m := range packet.Messages() {
			rrw.proxy.messages <- m
		}
	}
	return len(p), nil
}
