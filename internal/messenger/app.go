package messenger

import (
	"fmt"
	"strings"

	"github.com/woka00/Messenger-NEMAX/internal/httpapi"
	"github.com/woka00/Messenger-NEMAX/internal/sessions"
	"github.com/woka00/Messenger-NEMAX/internal/tcpserver"
	"github.com/woka00/Messenger-NEMAX/internal/users"
	"github.com/woka00/Messenger-NEMAX/internal/ws"
)

// Run initializes in-memory storage and starts HTTP and TCP servers.
func Run(tcpAddr, httpAddr string) error {
	if err := users.InitDefaults(); err != nil {
		return fmt.Errorf("initialize default users: %w", err)
	}
	sessions.StartCleanup()

	hub := ws.NewHub()
	go hub.Run()

	httpSrv := httpapi.NewServer(hub)
	go func() {
		fmt.Println("Web UI доступен по адресу", "http://"+trimHost(httpAddr))
		if err := httpSrv.Run(httpAddr); err != nil {
			fmt.Println("HTTP server error:", err)
		}
	}()

	fmt.Println("Messenger TCP server listening on", tcpAddr)
	return tcpserver.Run(tcpAddr)
}

func trimHost(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "localhost" + addr
	}
	return addr
}
