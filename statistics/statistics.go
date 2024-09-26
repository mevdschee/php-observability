package statistics

import (
	"compress/gzip"
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

func New() *Statistics {
	return NewWithBuckets([]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10})
}

func NewWithBuckets(buckets []float64) *Statistics {
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

func (s *Statistics) Add(name string, labelName string, labelValue string, duration float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := name + "|" + labelName
	ss, exists := s.Names[key]
	if !exists {
		ss = StatisticSet{
			Counters:  map[string]uint64{},
			Durations: map[string]float64{},
			Buckets:   map[string]uint64{},
		}
		s.Names[key] = ss
	}
	ss.Counters[labelValue]++
	ss.Durations[labelValue] += duration
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
	(*writer).Header().Set("Content-Encoding", "gzip")
	gw := gzip.NewWriter((*writer))
	defer gw.Close()
	names := make([]string, 0, len(s.Names))
	for name := range s.Names {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		ss := s.Names[name]
		parts := strings.SplitN(name, "|", 2)
		metricName := parts[0]
		labelName := parts[1]
		// counters
		gw.Write([]byte("# TYPE " + metricName + "_seconds summary\n"))
		gw.Write([]byte("# UNIT " + metricName + "_seconds seconds\n"))
		gw.Write([]byte("# HELP " + metricName + "_seconds A summary of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
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
			gw.Write([]byte(metricName + "_seconds_count{" + labelName + "=" + strconv.Quote(k) + "} " + strconv.FormatUint(c, 10) + "\n"))
			s := ss.Durations[k]
			sum += s
			gw.Write([]byte(metricName + "_seconds_sum{" + labelName + "=" + strconv.Quote(k) + "} " + strconv.FormatFloat(s, 'f', 3, 64) + "\n"))
		}
		// totals
		gw.Write([]byte("# TYPE " + metricName + "_total_seconds histogram\n"))
		gw.Write([]byte("# UNIT " + metricName + "_total_seconds seconds\n"))
		gw.Write([]byte("# HELP " + metricName + "_total_seconds A histogram of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
		for _, b := range s.Buckets {
			v := ss.Buckets[b.Name]
			gw.Write([]byte(metricName + "_total_seconds_bucket{le=" + strconv.Quote(b.Name) + "} " + strconv.FormatUint(v, 10) + "\n"))
		}
		gw.Write([]byte(metricName + "_total_seconds_sum " + strconv.FormatFloat(sum, 'f', 3, 64) + "\n"))
		gw.Write([]byte(metricName + "_total_seconds_count " + strconv.FormatUint(count, 10) + "\n"))
	}
	gw.Write([]byte("# EOF\n"))
}

func (s *Statistics) AddStatistics(s2 *Statistics) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for key, value := range s2.Names {
		ss, exists := s.Names[key]
		if !exists {
			s.Names[key] = value
		} else {
			for k, v := range value.Counters {
				ss.Counters[k] += v
			}
			for k, v := range value.Durations {
				ss.Durations[k] += v
			}
			for k, v := range value.Buckets {
				ss.Buckets[k] += v
			}
		}
	}
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
