package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

// stdin/stdout tcp
// addr "127.0.0.1:9000"
func runClient(addr string) error {
	fmt.Println("Подключаюсь к серверу", addr)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	go readFromServer(conn)

	reader := bufio.NewReader(os.Stdin)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Ошибка чтения ввода:", err)
			return nil
		}
		_, err = conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Ошибка отправки на сервер:", err)
			return nil
		}
	}
}

func readFromServer(conn net.Conn) {
	serverReader := bufio.NewReader(conn)
	for {
		line, err := serverReader.ReadString('\n')
		if err != nil {
			fmt.Println("Соединение закрыто сервером:", err)
			os.Exit(0)
		}
		fmt.Print(line)
	}
}
