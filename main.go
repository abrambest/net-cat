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

func writeHistory(userName string, clientConn net.Conn) {
	readhistory, err := os.ReadFile("history.txt")
	if err != nil {
		log.Println(err)
	}
	if len(string(readhistory)) != 0 {
		fmt.Fprintf(clientConn, string(readhistory))
		fmt.Fprintf(clientConn, "[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), userName)

	} else {
		fmt.Fprintf(clientConn, "[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), userName)
	}
}

func cover(name, text string) massageData {
	return massageData{
		Name:    name,
		Massage: text,
	}
}

func checkNames(name string) bool {
	for _, names := range usersMap {
		if name == names {
			return false
		}
	}
	return true
}

// func historyWrite()

func checkNamesFonts(name string) bool {
	for _, s := range name {
		if (s < '0' || s > '9') && (s < 'A' || s > 'Z') && (s < 'a' || s > 'z') {
			return false
		}
	}

	return true
}

func handle(clientConn net.Conn) {
	defer clientConn.Close()
	var getNameUser string

	welcome, err := os.ReadFile("welcome.txt")
	if err != nil {
		log.Println(err)
	}
	clientConn.Write(welcome)

	for {
		clientConn.Write([]byte("[ENTER YOUR NAME]: "))

		name := bufio.NewReader(clientConn)
		getNameUser, err = name.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		if len(usersMap) > 9 {
			clientConn.Write([]byte("Chat is full, try later...\n"))
			continue
		} else if !checkNames(strings.TrimSpace(getNameUser)) {
			clientConn.Write([]byte("Name is busy\n"))
			continue
		} else if !checkNamesFonts(strings.TrimSpace(getNameUser)) {
			clientConn.Write([]byte("Incorrect username\n"))
			continue
		} else if strings.TrimSpace(getNameUser) == "\n" {
			continue
		} else {
			break
		}
	}

	userName := strings.TrimSpace(getNameUser)

	mu.Lock()
	usersMap[clientConn] = userName
	info <- cover(userName, "has joined our chat...")
	mu.Unlock()

	clientScaner := bufio.NewScanner(clientConn)

	mu.Lock()
	writeHistory(userName, clientConn)
	mu.Unlock()

	for clientScaner.Scan() {

		fmt.Fprintf(clientConn, "[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), userName)
		scanTxt := strings.TrimSpace(clientScaner.Text())

		if scanTxt == "" {
			fmt.Fprintf(clientConn, "Error - sent an empty message...\n[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), userName)
			continue
		} else {
			data := cover(userName, scanTxt)

			mu.Lock()
			historyAdd(data) // historyChan <- data
			mu.Unlock()

			letter <- data

		}

	}
	mu.Lock()
	delete(usersMap, clientConn)
	info <- cover(userName, "has left our chat...")
	mu.Unlock()
}

func historyAdd(h massageData) {
	options := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	history, err := os.OpenFile("history.txt", options, 0666)
	defer history.Close()
	if err != nil {
		log.Println(err)
	}
	fmt.Fprintf(history, "[%s][%s]:%s\n", time.Now().Format("2006-1-2 15:4:5"), h.Name, h.Massage)
}

func postMan() {
	for {
		select {
		case letter := <-letter:
			for conn, user := range usersMap {
				if user == letter.Name {
					continue
				}
				fmt.Fprintf(conn, "\n[%s][%s]:%s\n[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), letter.Name, letter.Massage, time.Now().Format("2006-1-2 15:4:5"), user)
			}
		case info := <-info:
			for conn, user := range usersMap {
				if user == info.Name {
					continue
				}
				fmt.Fprintf(conn, "\n%s %s\n[%s][%s]:", info.Name, info.Massage, time.Now().Format("2006-1-2 15:4:5"), user)
			}

		}
	}
}

func main() {
	port := ""
	if len(os.Args) > 2 {
		log.Println("[USAGE]: ./TCPChat $port")
		return
	}
	if len(os.Args) == 2 {
		port = os.Args[1]
	} else {
		port = "8989"
	}

	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println(err, "-> Please input port: \"Example: go run . 8080\"")
		return
	}
	fmt.Println("Listening on the port :" + port)

	defer server.Close()

	history, err := os.OpenFile("history.txt", os.O_TRUNC, 0666)
	if err != nil {
		log.Println(err)
	}
	defer history.Close()

	for {

		client, err := server.Accept()
		if err != nil {
			log.Println(err.Error())
		}

		go handle(client)
		go postMan()

	}
}
