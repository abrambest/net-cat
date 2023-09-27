package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type massageData struct {
	Name    string
	Massage string
}

var (
	usersMap    = make(map[net.Conn]string)
	mu          sync.Mutex
	letter      = make(chan massageData)
	info        = make(chan massageData)
	historyChan = make(chan massageData)
)

func cover(name, txt string) massageData {
	return massageData{
		Name:    name,
		Massage: txt,
	}

}
func writeHistory(name string, conn net.Conn) {
	history, err := os.ReadFile("history.txt")
	if err != nil {
		fmt.Println("error read history")
		return
	}

	if len(history) != 0 {
		conn.Write(history)
		fmt.Fprintf(conn, "[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), name)

	} else {
		return

	}

}
func historyAdd(data massageData) {
	option := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	history, err := os.OpenFile("history.txt", option, 0666)
	defer history.Close()
	if err != nil {
		log.Println("error add history to file", err)
		return
	}
	fmt.Fprintf(history, "[%s][%s]:%s\n", time.Now().Format("2006-1-2 15:4:5"), data.Name, data.Massage)

}

func chackNameFonts(name string) bool {
	for _, s := range name {
		if (s < '0' || s > '9') && (s < 'A' || s > 'Z') && (s < 'a' || s > 'z') {
			return false
		}
	}
	return true

}

func checkNames(name string) bool {
	for _, nameUser := range usersMap {
		if name == nameUser {
			return false
		}
	}
	return true

}

func postMan() {
	for {
		select {
		case letter := <-letter:
			mu.Lock()
			for conn, user := range usersMap {
				if letter.Name == user {
					continue
				}
				fmt.Fprintf(conn, "\n[%s][%s]:%s\n[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), letter.Name, letter.Massage, time.Now().Format("2006-1-2 15:4:5"), user)
			}
			mu.Unlock()

		case info := <-info:
			mu.Lock()
			for conn, user := range usersMap {
				if info.Name == user {
					continue
				}
				fmt.Fprintf(conn, "\n%s %s\n[%s][%s]:", info.Name, info.Massage, time.Now().Format("2006-1-2 15:4:5"), user)
			}
			mu.Unlock()
		}
	}

}

func handle(clientConn net.Conn) {

	defer clientConn.Close()
	var getUserName string
	welcome, err := os.ReadFile("welcome.txt")
	if err != nil {
		log.Println("error read file welcom.txt, ", err)
	}
	clientConn.Write(welcome)
	for {
		clientConn.Write([]byte("ENTER YOUR NAME: "))
		name := bufio.NewReader(clientConn)
		getUserName, err = name.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}
		if len(usersMap) > 9 {
			clientConn.Write([]byte("Chat is full, please try leater...\n"))
			continue
		} else if strings.TrimSpace(getUserName) == "" {
			clientConn.Write([]byte("Please enter your name"))
			continue
		} else if !checkNames(strings.TrimSpace(getUserName)) {
			clientConn.Write([]byte("Name is busy\n"))
			continue
		} else if !chackNameFonts(strings.TrimSpace(getUserName)) {
			clientConn.Write([]byte("Incorrect username\n"))
			continue
		} else if strings.TrimSpace(getUserName) == "\n" {
			continue
		} else {
			break
		}
	}
	userName := strings.TrimSpace(getUserName)
	mu.Lock()
	usersMap[clientConn] = userName
	info <- cover(userName, "has joined our chat...")
	mu.Unlock()

	clientScanner := bufio.NewScanner(clientConn)
	mu.Lock()
	writeHistory(userName, clientConn)
	mu.Unlock()

	for {

		fmt.Fprintf(clientConn, "rr[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), userName)
		ok := clientScanner.Scan()
		if !ok {
			break
		}

		scanTxt := strings.TrimSpace(clientScanner.Text())

		if scanTxt == "" {
			fmt.Fprintf(clientConn, "Error - sent an empty message...\n[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), userName)
			continue
		} else {
			clientConn.Write([]byte("\033[1A"))
			clientConn.Write([]byte("\033[K"))
			fmt.Fprintf(clientConn, "WWW[%s][%s]: %s\n", time.Now().Format("2006-1-2 15:4:5"), userName, scanTxt)
			data := cover(userName, scanTxt)
			mu.Lock()
			historyAdd(data)
			mu.Unlock()
			letter <- data
		}
	}
	mu.Lock()
	delete(usersMap, clientConn)
	info <- cover(userName, "has left our chat...")
	mu.Unlock()

}

func main() {
	port := "8989"
	if len(os.Args) > 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}
	if len(os.Args) == 2 {
		port = os.Args[1]
	}
	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println(err, err, "-> Please input port: \"Example: go run . 8080\"")
	}
	fmt.Println("listening on the port: " + port)
	defer server.Close()

	history, err := os.OpenFile("history.txt", os.O_TRUNC, 0666)
	if err != nil {
		log.Println("error creat or open history file", err)
		return
	}
	defer history.Close()
	for {
		client, err := server.Accept()
		if err != nil {
			log.Println("error accepting net.Conn", err)
		}

		go handle(client)
		go postMan()

	}

}
