package tcpserver

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"olimps/internal/messages"
	"olimps/internal/ratelimit"
	"olimps/internal/users"
)

// поднимает tcp мессенджер
func Run(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	reader := bufio.NewScanner(conn)
	currentUser := ""

	writeLine(writer, "Добро пожаловать в текстовый мессенджер.")
	writeLine(writer, "Команды: LOGIN, SEND, INBOX, USERS, LOGOUT, HELP, QUIT")
	writeLine(writer, "Пример: LOGIN alice 1111")

	for {
		writeLine(writer, "")
		writeLine(writer, "> ")

		if !reader.Scan() {
			return
		}
		line := strings.TrimSpace(reader.Text())
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 3)
		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "LOGIN":
			if len(parts) < 3 {
				writeLine(writer, "Используйте: LOGIN <login> <password>")
				continue
			}

			connID := conn.RemoteAddr().String()

			if !ratelimit.CheckRateLimit(connID) {
				writeLine(writer, "Превышен лимит попыток. Попробуйте позже.")
				continue
			}
			if blocked, until := ratelimit.CheckLoginBlocked(connID); blocked {
				remaining := time.Until(until).Round(time.Second)
				writeLine(writer, fmt.Sprintf("IP временно заблокирован. Попробуйте через %v", remaining))
				continue
			}

			login := parts[1]
			password := parts[2]
			psha := sha256.Sum256([]byte(password))
			passwordBase64 := base64.StdEncoding.EncodeToString(psha[:])
			if !users.CheckCredentials(login, passwordBase64) {
				ratelimit.RecordFailed(connID)
				writeLine(writer, "Неверный логин или пароль")
				continue
			}

			ratelimit.RecordSuccess(connID)
			currentUser = login
			writeLine(writer, "Успешный вход: "+currentUser)

		case "USERS":
			writeLine(writer, "Зарегистрированные пользователи:")
			for _, u := range users.List() {
				writeLine(writer, " - "+u)
			}

		case "SEND":
			if currentUser == "" {
				writeLine(writer, "Сначала выполните LOGIN")
				continue
			}
			if len(parts) < 3 {
				writeLine(writer, "Используйте: SEND <to_login> <text>")
				continue
			}

			toLoginAndMaybeText := strings.SplitN(line, " ", 3)
			if len(toLoginAndMaybeText) < 3 {
				writeLine(writer, "Используйте: SEND <to_login> <text>")
				continue
			}
			toLogin := toLoginAndMaybeText[1]
			text := toLoginAndMaybeText[2]

			if !users.Exists(toLogin) {
				writeLine(writer, "Неизвестный получатель: "+toLogin)
				continue
			}

			msg := messages.Message{
				From: currentUser,
				To:   toLogin,
				Text: text,
				Time: time.Now().UnixNano(),
			}
			messages.Add(toLogin, msg)
			writeLine(writer, "Сообщение отправлено пользователю "+toLogin)

		case "INBOX":
			if currentUser == "" {
				writeLine(writer, "Сначала выполните LOGIN")
				continue
			}
			msgs := messages.Inbox(currentUser)
			if len(msgs) == 0 {
				writeLine(writer, "Новых сообщений нет.")
				continue
			}
			writeLine(writer, "Сообщения:")
			for i, m := range msgs {
				writeLine(writer, fmt.Sprintf("%d) от %s: %s", i+1, m.From, m.Text))
			}

		case "LOGOUT":
			if currentUser == "" {
				writeLine(writer, "Вы не авторизованы.")
				continue
			}
			currentUser = ""
			writeLine(writer, "Вы вышли из аккаунта.")

		case "HELP":
			writeLine(writer, "Команды:")
			writeLine(writer, " LOGIN <login> <password>  - вход в систему")
			writeLine(writer, " SEND <to_login> <text>    - отправить сообщение")
			writeLine(writer, " INBOX                     - показать новые сообщения")
			writeLine(writer, " USERS                     - список пользователей")
			writeLine(writer, " LOGOUT                    - выйти из аккаунта")
			writeLine(writer, " QUIT                      - закрыть соединение")

		case "QUIT", "EXIT":
			writeLine(writer, "Пока!")
			return

		default:
			writeLine(writer, "Неизвестная команда. Введите HELP для справки.")
		}
	}
}

func writeLine(w *bufio.Writer, s string) {
	_, _ = w.WriteString(s + "\n")
	_ = w.Flush()
}
