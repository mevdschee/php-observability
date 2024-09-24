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

func handleConn(conn net.Conn) {
	defer conn.Close()
	scan := bufio.NewScanner(conn)
	for scan.Scan() {
		input := scan.Text()
		var fields []any
		err := json.Unmarshal([]byte(input), &fields)
		if err != nil || len(fields) != 4 {
			log.Printf("malformed input: %v\n", input)
			continue
		}
		metricName, ok := fields[0].(string)
		if !ok {
			log.Printf("malformed input at pos 0 (metricName)")
			continue
		}
		tagName, ok := fields[1].(string)
		if !ok {
			log.Printf("malformed input at pos 1 (tagName)")
			continue
		}
		tagValue, ok := fields[2].(string)
		if !ok {
			log.Printf("malformed input at pos 2 (tagValue)")
			continue
		}
		duration, ok := fields[3].(float64)
		if !ok {
			log.Printf("malformed input at pos 3 (duration)")
			continue
		}
		stats.Add(metricName, tagName, tagValue, duration)
		log.Printf("received input: %v", fields)
	}
}
