package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/mevdschee/php-observability/metrics"
)

var stats = metrics.New()

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile := flag.String("memprofile", "", "write mem profile to file")
	urlsToScrape := flag.String("scrape", "", "comma seperated list of URLs to scrape for Gob metrics")
	scrapeEvery := flag.Duration("every", 5*time.Second, "seconds to wait between scrape requests")
	listenAddress := flag.String("listen", "localhost:7777", "address to listen for high frequent events over TCP")
	metricsAddress := flag.String("metrics", ":8080", "address to listen for Prometheus metric scraper over HTTP")
	binaryAddress := flag.String("binary", ":9999", "address to listen for Gob metric scraper over HTTP")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	go serve(*memprofile, *metricsAddress)
	go serveGob(*binaryAddress)
	go scrapeUrlsEvery(*urlsToScrape, *scrapeEvery, stats)
	logListener(*listenAddress)
}

func serve(memprofile, metricsAddress string) {
	err := http.ListenAndServe(metricsAddress, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		stats.Write(&writer)
		if memprofile != "" {
			f, err := os.Create(memprofile)
			if err != nil {
				log.Fatal(err)
			}
			pprof.WriteHeapProfile(f)
			f.Close()
		}
	}))
	log.Fatal(err)
}

func serveGob(metricsAddress string) {
	err := http.ListenAndServe(metricsAddress, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		stats.WriteGob(&writer)
	}))
	log.Fatal(err)
}

func getMetrics(url string) (*metrics.Metrics, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status: %v", resp.StatusCode)
	}
	s := metrics.New()
	err = s.ReadGob(resp)
	if err != nil {
		return nil, fmt.Errorf("http read body: %v", err)
	}
	return s, nil
}

func scrapeUrlsEvery(urlsToScrape string, scrapeEvery time.Duration, stats *metrics.Metrics) {
	if len(urlsToScrape) == 0 {
		return
	}
	urls := strings.Split(urlsToScrape, ",")
	ticker := time.NewTicker(scrapeEvery)
	for range ticker.C {
		scrapeUrls(urls, stats)
	}
}

func scrapeUrls(urls []string, stats *metrics.Metrics) {
	ch := make(chan *metrics.Metrics)
	for _, url := range urls {
		go func(url string) {
			log.Printf("get: %v", url)
			s, err := getMetrics(url)
			if err != nil {
				log.Printf("scrape error: %v", err)
			}
			ch <- s
		}(strings.TrimSpace(url))
	}
	for range urls {
		s := <-ch
		if s != nil {
			if stats == nil {
				stats = s
			} else {
				stats.AddMetrics(s)
			}
		}
	}
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
