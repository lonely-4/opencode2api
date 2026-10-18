package gateway

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestSourceDialBindsLocalIPv6(t *testing.T) {
	ln, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	seen := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			seen <- err.Error()
			return
		}
		defer conn.Close()
		seen <- conn.RemoteAddr().String()
	}()

	dial := sourceDialContext(net.ParseIP("::1"), 2*time.Second)
	conn, err := dial(context.Background(), "tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	conn.Close()

	remote := <-seen
	if !strings.HasPrefix(remote, "[::1]:") {
		t.Fatalf("accepted remote %q, want a connection from ::1", remote)
	}
}

func TestSourceDialRejectsIPv4(t *testing.T) {
	dial := sourceDialContext(net.ParseIP("::1"), time.Second)
	if _, err := dial(context.Background(), "tcp", "192.0.2.1:80"); err == nil {
		t.Fatal("dial to an IPv4 host succeeded")
	}
}

func TestProxyClientRotatesSources(t *testing.T) {
	p := &proxyTransport{clients: []*http.Client{{}, {}, {}}}
	seen := map[*http.Client]int{}
	for range 300 {
		seen[p.Client()]++
	}
	if len(seen) != 3 {
		t.Fatalf("random selection hit %d clients, want 3", len(seen))
	}
}
