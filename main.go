package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/mevdschee/php-observability/statistics"
)

var stats = statistics.New([]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10})

func main() {
	listenAddress := flag.String("listen", ":7777", "address to listen for high frequent events over TCP")
	metricsAddress := flag.String("metrics", ":4000", "address to listen for Prometheus metric scraper over HTTP")
	flag.Parse()
	go serve(*metricsAddress)
	logListener(*listenAddress)
}

func serve(metricsAddress string) {
	http.ListenAndServe(metricsAddress, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		stats.Write(&writer)
	}))
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
		fields := strings.Split(input, ":")
		duration, _ := strconv.ParseFloat(fields[3], 64)
		stats.Add(fields[0], fields[1], fields[2], duration)
		log.Printf("received input: %v", fields)
	}
}
