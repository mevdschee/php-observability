package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/mevdschee/php-observability/statistics"
)

var stats = statistics.NewDefault()

func main() {
	listenAddress := flag.String("listen", "localhost:7777", "address to listen for high frequent events over TCP")
	metricsAddress := flag.String("metrics", ":8080", "address to listen for Prometheus metric scraper over HTTP")
	binaryAddress := flag.String("binary", ":9999", "address to listen for Gob metric scraper over HTTP")
	flag.Parse()
	go serve(*metricsAddress)
	go serveGob(*binaryAddress)
	logListener(*listenAddress)
}

func serve(metricsAddress string) {
	err := http.ListenAndServe(metricsAddress, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		stats.Write(&writer)
	}))
	log.Fatal(err)
}

func serveGob(metricsAddress string) {
	err := http.ListenAndServe(metricsAddress, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		stats.WriteGob(&writer)
	}))
	log.Fatal(err)
}

func logListener(listenAddress string) {
	lis, err := net.Listen("tcp", listenAddress)
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
		var fields []string
		err := json.Unmarshal([]byte(input), &fields)
		if err != nil || len(fields) != 4 {
			log.Printf("malformed input: %v", input)
			continue
		}
		duration, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			log.Printf("malformed duration: %v", fields[3])
			continue
		}
		stats.Add(fields[0], fields[1], fields[2], duration)
		log.Printf("received input: %v", input)
	}
}
