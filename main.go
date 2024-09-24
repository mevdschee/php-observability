package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"

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
		log.Fatalf("failed to start listener: %v\n", err)
	}
	defer lis.Close()

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Printf("failed to accept conn: %v\n", err)
			continue
		}
		go handleConn(conn)
	}
}

type Metric struct {
	Key   []string `json:"k"`
	Value float64  `json:"v"`
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	scan := bufio.NewScanner(conn)
	for scan.Scan() {
		input := scan.Text()
		var metric Metric
		err := json.Unmarshal([]byte(input), &metric)
		if err != nil || len(metric.Key) != 3 {
			log.Printf("malformed input: %v\n", input)
			continue
		}
		stats.Add(metric.Key[0], metric.Key[1], metric.Key[2], metric.Value)
		log.Printf("received input: %v", input)
	}
}
