package statistics

import (
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Bucket struct {
	Name  string
	Value float64
}

type StatisticSet struct {
	Counters  map[string]uint64
	Durations map[string]float64
	Buckets   map[string]uint64
}

type Statistics struct {
	mutex   sync.Mutex
	Names   map[string]StatisticSet
	Buckets []Bucket
}

func NewDefault() *Statistics {
	return New([]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10})
}

func New(buckets []float64) *Statistics {
	s := Statistics{
		Names:   map[string]StatisticSet{},
		Buckets: []Bucket{},
	}
	sort.Float64s(buckets)
	for _, value := range buckets {
		name := fmt.Sprintf("%g", value)
		s.Buckets = append(s.Buckets, Bucket{name, value})
	}
	s.Buckets = append(s.Buckets, Bucket{"+Inf", math.MaxFloat64})
	return &s
}

func (s *Statistics) Add(name string, tagName string, tagValue string, duration float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := name + "|" + tagName
	ss, exists := s.Names[key]
	if !exists {
		ss = StatisticSet{
			Counters:  map[string]uint64{},
			Durations: map[string]float64{},
			Buckets:   map[string]uint64{},
		}
		s.Names[key] = ss
	}
	ss.Counters[tagValue]++
	ss.Durations[tagValue] += duration
	for i := len(s.Buckets) - 1; i >= 0; i-- {
		b := s.Buckets[i]
		if b.Value < duration {
			break
		}
		ss.Buckets[b.Name]++
	}
}

func (s *Statistics) Write(writer *http.ResponseWriter) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	names := make([]string, 0, len(s.Names))
	for name := range s.Names {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		ss := s.Names[name]
		parts := strings.SplitN(name, "|", 2)
		metricName := parts[0]
		tagName := parts[1]
		// counters
		(*writer).Write([]byte("# TYPE " + metricName + "_seconds summary\n"))
		(*writer).Write([]byte("# UNIT " + metricName + "_seconds seconds\n"))
		(*writer).Write([]byte("# HELP " + metricName + "_seconds A summary of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
		keys := make([]string, 0, len(ss.Counters))
		for key := range ss.Counters {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		count := uint64(0)
		sum := float64(0)
		for _, k := range keys {
			c := ss.Counters[k]
			count += c
			(*writer).Write([]byte(metricName + "_seconds_count{" + tagName + "=" + strconv.Quote(k) + "} " + strconv.FormatUint(c, 10) + "\n"))
			s := ss.Durations[k]
			sum += s
			(*writer).Write([]byte(metricName + "_seconds_sum{" + tagName + "=" + strconv.Quote(k) + "} " + strconv.FormatFloat(s, 'f', 3, 64) + "\n"))
		}
		// totals
		(*writer).Write([]byte("# TYPE " + metricName + "_total_seconds histogram\n"))
		(*writer).Write([]byte("# UNIT " + metricName + "_total_seconds seconds\n"))
		(*writer).Write([]byte("# HELP " + metricName + "_total_seconds A histogram of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
		for _, b := range s.Buckets {
			v := ss.Buckets[b.Name]
			(*writer).Write([]byte(metricName + "_total_seconds_bucket{le=" + strconv.Quote(b.Name) + "} " + strconv.FormatUint(v, 10) + "\n"))
		}
		(*writer).Write([]byte(metricName + "_total_seconds_sum " + strconv.FormatFloat(sum, 'f', 3, 64) + "\n"))
		(*writer).Write([]byte(metricName + "_total_seconds_count " + strconv.FormatUint(count, 10) + "\n"))
	}
	(*writer).Write([]byte("# EOF\n"))
}

func (s *Statistics) WriteGob(writer *http.ResponseWriter) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	enc := gob.NewEncoder((*writer))
	err := enc.Encode(s)
	if err != nil {
		log.Fatal("encode:", err)
	}
}
