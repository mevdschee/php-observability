package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/mevdschee/php-observability/statistics"
)

var stats = statistics.New()

func main() {
	go serve()
	logListener()
}

func serve() {
	http.ListenAndServe(":4000", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		stats.Write(&writer)
	}))
}

func logListener() {
	lis, err := net.Listen("tcp", ":7777")
	if err != nil {
		log.Fatalf("failed to start listener: %v", err)
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("failed to accept conn: %v", err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	scan := bufio.NewScanner(conn)
	for scan.Scan() {
		input := scan.Text()
		fields := strings.Split(input, ":")
		duration, _ := strconv.ParseFloat(fields[3], 64)
		stats.Add(fields[0], fields[1], fields[2], duration)
		log.Printf("received input: %v", fields)
	}
}
