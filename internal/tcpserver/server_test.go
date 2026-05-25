package tcpserver

import (
	"net"
	"testing"
)

func TestRemoteIPDoesNotIncludeClientPort(t *testing.T) {
	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 54321}
	if got := remoteIP(addr); got != "127.0.0.1" {
		t.Fatalf("remoteIP() = %q, want %q", got, "127.0.0.1")
	}
}
