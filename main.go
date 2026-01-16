package main

import (
	"flag"
	"fmt"
	"os"

	"olimps/internal/messenger"
)

// Zapusk servera ili centa bez lishnego
//
//	go run . -mode=server -addr=":9000"
//	go run . -mode=client -addr="127.0.0.1:9000"
func main() {
	mode := flag.String("mode", "server", "Режим работы: server или client")
	addr := flag.String("addr", ":9000", "Адрес TCP сервера ':9000' или 'host:port'")
	flag.Parse()

	switch *mode {
	case "server":
		if err := messenger.Run(*addr, ":8080"); err != nil {
			fmt.Println("Ошибка запуска сервера:", err)
			os.Exit(1)
		}
	case "client":
		if err := runClient(*addr); err != nil {
			fmt.Println("Ошибка клиента:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Неизвестный режим. Используйте -mode=server или -mode=client")
		os.Exit(1)
	}
}
