package httpx

import (
	"context"
	"io"
	"net"
	"testing"
)

func TestBunDialUsesDialContext(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	accepted := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		io.Copy(io.Discard, conn)
		conn.Close()
		close(accepted)
	}()

	called := false
	tr, err := NewBunTransport(BunOptions{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			called = true
			if addr != "example.invalid:80" {
				t.Errorf("dial addr = %q", addr)
			}
			return (&net.Dialer{}).DialContext(ctx, "tcp", ln.Addr().String())
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := tr.dial(context.Background(), "http", "example.invalid:80", "example.invalid")
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()
	if !called {
		t.Fatal("Bun dial ignored DialContext")
	}
	<-accepted
}
