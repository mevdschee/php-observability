package statistics

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type StatisticSet struct {
	counters  map[string]uint64
	durations map[string]float64
}

type Statistics struct {
	mutex sync.Mutex
	names map[string]StatisticSet
}

func New() *Statistics {
	s := Statistics{
		names: map[string]StatisticSet{},
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
		}
		s.names[key] = ss
	}
	ss.counters[tag]++
	ss.durations[tag] += val
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
		for _, k := range keys {
			v := ss.counters[k]
			(*writer).Write([]byte(metricName + "_count{" + tagName + "=\"" + k + "\"} " + strconv.FormatUint(v, 10) + "\n"))
		}
		keys = []string{}
		for key := range ss.durations {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := ss.durations[k]
			(*writer).Write([]byte(metricName + "_seconds{" + tagName + "=\"" + k + "\"} " + strconv.FormatFloat(v, 'f', 3, 64) + "\n"))
		}
	}
}
