package main

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/onunez-g/go-chat/chat"
	"github.com/onunez-g/go-chat/utils"
)

func main() {
	s := chat.NewServer()
	go s.Run()

	port := findPort()
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Unable to start server: %s", err.Error())
	}
	defer listener.Close()

	log.Println("started server on " + port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Unable to accept connection: %s", err.Error())
			continue
		}
		go s.NewClient(conn)
	}
}

func findPort() string {
	args := strings.Join(os.Args[1:], " ")
	if strings.Contains(args, "-p") {
		index := utils.FindIndex(os.Args, "-p")
		return ":" + os.Args[index+1]
	} else if strings.Contains(args, "--port") {
		index := utils.FindIndex(os.Args, "--port")
		return ":" + os.Args[index+1]
	}
	return ":5000"
}
