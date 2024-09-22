package statistics

import (
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Boundary struct {
	name  string
	value float64
}

type StatisticSet struct {
	counters  map[string]uint64
	durations map[string]float64
	buckets   map[string]uint64
}

type Statistics struct {
	mutex      sync.Mutex
	names      map[string]StatisticSet
	boundaries []Boundary
}

func New() *Statistics {
	s := Statistics{
		names: map[string]StatisticSet{},
		boundaries: []Boundary{
			{"0.001", 0.001},
			{"0.01", 0.01},
			{"0.1", 0.1},
			{"1", 1},
			{"10", 10},
			{"100", 100},
			{"+Inf", math.MaxFloat64},
		},
	}
	return &s
}

func (s *Statistics) Add(name string, tagName string, tag string, val float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	key := name + "|" + tagName
	ss, exists := s.names[key]
	if !exists {
		ss = StatisticSet{
			counters:  map[string]uint64{},
			durations: map[string]float64{},
			buckets:   map[string]uint64{},
		}
		s.names[key] = ss
	}
	ss.counters[tag]++
	ss.durations[tag] += val
	for _, b := range s.boundaries {
		if b.value > val {
			break
		}
		ss.buckets[b.name]++
	}
}

func (s *Statistics) Write(writer *http.ResponseWriter) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var names []string
	for name := range s.names {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		ss := s.names[name]
		parts := strings.SplitN(name, "|", 2)
		metricName := parts[0]
		tagName := parts[1]
		var keys []string
		for key := range ss.counters {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		count := uint64(0)
		for _, k := range keys {
			v := ss.counters[k]
			count += v
			(*writer).Write([]byte(metricName + "_count{" + tagName + "=\"" + k + "\"} " + strconv.FormatUint(v, 10) + "\n"))
		}
		keys = []string{}
		for key := range ss.durations {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		sum := float64(0)
		for _, k := range keys {
			v := ss.durations[k]
			sum += v
			(*writer).Write([]byte(metricName + "_seconds{" + tagName + "=\"" + k + "\"} " + strconv.FormatFloat(v, 'f', 3, 64) + "\n"))
		}
		for _, b := range s.boundaries {
			v := ss.buckets[b.name]
			(*writer).Write([]byte(metricName + "_seconds_bucket{le=\"" + b.name + "\"} " + strconv.FormatUint(v, 10) + "\n"))
		}
		(*writer).Write([]byte(metricName + "_seconds_count " + strconv.FormatUint(count, 10) + "\n"))
		(*writer).Write([]byte(metricName + "_seconds_total " + strconv.FormatFloat(sum, 'f', 3, 64) + "\n"))
	}
}
