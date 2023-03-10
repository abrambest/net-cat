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

type StructUsers struct {
	Addr net.Conn
	Name string
}

var (
	usersMap = make(map[string]StructUsers)
	mu       sync.Mutex
)

func checkNames(name string) bool {
	for names := range usersMap {
		if name == names {
			return false
		}
	}
	return true
}

func checkNamesFonts(name string) bool {
	for _, s := range name {
		if (s < '0' || s > '9') && (s < 'A' || s > 'Z') && (s < 'a' || s > 'z') {
			fmt.Println("false")
			return false
		}
	}
	fmt.Println("true")
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

	user := StructUsers{
		Addr: clientConn,
		Name: userName,
	}
	mu.Lock()
	usersMap[user.Name] = user
	mu.Unlock()

	// fmt.Printf("Name user: %s, RemoteAddr: %s\n", userName, clientConn.RemoteAddr())
	clientScaner := bufio.NewScanner(clientConn)

	fmt.Fprintf(clientConn, "[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), userName)

	for clientScaner.Scan() {
		fmt.Fprintf(clientConn, "[%s][%s]:", time.Now().Format("2006-1-2 15:4:5"), userName)
		scanTxt := strings.TrimSpace(clientScaner.Text())

		// line, err := b.ReadBytes('\n')
		// if err != nil {
		// 	log.Println(err, user.Name)
		// 	return
		// }
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

	fmt.Println("Listening on the port :" + port)

	server, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Println(err, "-> Please input port: \"Example: go run . 8080\"")
		return
	}
	defer server.Close()

	history, err := os.Create("history.txt")
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

	}
}
