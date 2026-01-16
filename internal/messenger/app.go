package messenger

import (
	"fmt"
	"strings"

	"olimps/internal/httpapi"
	"olimps/internal/sessions"
	"olimps/internal/tcpserver"
	"olimps/internal/users"
	"olimps/internal/ws"
)

// Run стартует http+tcp и инициализирует все
func Run(tcpAddr, httpAddr string) error {
	users.InitDefaults()
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
