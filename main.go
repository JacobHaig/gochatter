package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	argparse "github.com/akamensky/argparse"
	"net"
	"os"
)

var messageChan = make(chan []byte)
var connections []*net.TCPConn

// gochatter is a chatting application.
// start as either a server with -s or connect via an IP with -a "0.0.0.0"
// if your a client, add the -u for the user name i.e. -u Wis

func main() {
	// Create new parser object
	parser := argparse.NewParser("Chatter", "Chatting application")

	username := parser.String("u", "username", &argparse.Options{Required: false, Help: "Username.", Default: "Wis"})
	address := parser.String("a", "address", &argparse.Options{Required: false, Help: "Sets the ip address. Default : 127.0.0.1", Default: "127.0.0.1"})
	port := parser.String("p", "port", &argparse.Options{Required: false, Help: "Sets the port. Default : 23432", Default: "23432"})
	isServer := parser.Flag("s", "server", &argparse.Options{Required: false, Help: "When flagged, server will start instead of client"})


	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	// Print the collected arg string
	fmt.Printf(
		"IP:%s Port:%s Server:%t Username:%s",
		*address,
		*port,
		*isServer,
		*username,
	)

	if *isServer {
		setupServer(port)
	} else {
		client(username, address, port)
	}
}

func setupServer(port *string) {
	// Resolve IP address
	addr, err := net.ResolveTCPAddr("tcp4", "0.0.0.0"+":"+*port)
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		panic(err)
	}

	go replicateMessages()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			panic(err)
		}

		connections = append(connections, conn)
		go server(conn)
	}
}

func replicateMessages() {
	// Write Connection
	for {
		bytes := <-messageChan

		for _, conn := range connections {
			WriteToConnection(conn, &bytes)
		}
	}
}

func server(conn *net.TCPConn) {
	// Read Connection
	for {
		data, err := ReadFromConnection(conn)
		if err != nil {
			conn.Close()
			return
		}

		messageChan <- data
		message := FromBytes(data)

		println(message.ToString())
	}
}

func client(username *string, address *string, port *string) {
	// Resolve IP address
	addr, err := net.ResolveTCPAddr("tcp4", *address+":"+*port)
	if err != nil {
		panic(err)
	}

	// Connect to server via address
	conn, err := net.DialTCP("tcp4", nil, addr)
	if err != nil {
		panic(err)
	}

	// Write Connection
	go func(conn *net.TCPConn) {
		reader := bufio.NewReader(os.Stdin)
		for {
			data, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			message := Message{
				Username: *username,
				Content:  data[:len(data)-1], // Remove "\n" from tale
			}

			bytes := message.ToBytes()

			WriteToConnection(conn, bytes)
		}
	}(conn)

	// Read Connection
	func(conn *net.TCPConn) {
		for {
			b, err := ReadFromConnection(conn)
			if err != nil {
				return
			}

			message := FromBytes(b)

			println(message.ToString())
		}
	}(conn)
}

func ReadFromConnection(conn *net.TCPConn) ([]byte, error) {
	bsize := make([]byte, 8)
	_, err1 := conn.Read(bsize)
	if err1 != nil {
		return nil, err1
	}
	size := binary.LittleEndian.Uint64(bsize)

	data := make([]byte, size)
	_, err2 := conn.Read(data)
	if err2 != nil {
		return nil, err2
	}

	return data, nil
}

func WriteToConnection(conn *net.TCPConn, data *[]byte) error {
	size := uint64(len(*data))

	bsize := make([]byte, 8)
	binary.PutUvarint(bsize, size)

	_, err1 := conn.Write(bsize)
	if err1 != nil {
		return err1
	}

	_, err2 := conn.Write(*data)
	if err2 != nil {
		return err2
	}

	return nil
}
