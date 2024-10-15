package metrics

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
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

type MetricSet struct {
	Counters       map[string]uint64
	DurationCounts map[string]uint64
	DurationSums   map[string]float64
	Buckets        map[string]uint64
}

type Metrics struct {
	mutex   sync.Mutex
	Names   map[string]MetricSet
	Buckets []Bucket
}

func New() *Metrics {
	return NewWithBuckets([]float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10})
}

func NewWithBuckets(buckets []float64) *Metrics {
	m := Metrics{
		Names:   map[string]MetricSet{},
		Buckets: []Bucket{},
	}
	sort.Float64s(buckets)
	for _, value := range buckets {
		name := fmt.Sprintf("%g", value)
		m.Buckets = append(m.Buckets, Bucket{name, value})
	}
	return &m
}

func (m *Metrics) Inc(name string, labelName string, labelValue string, delta uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	key := name + "|" + labelName
	ms, exists := m.Names[key]
	if !exists {
		ms = MetricSet{
			Counters:       map[string]uint64{},
			DurationCounts: map[string]uint64{},
			DurationSums:   map[string]float64{},
			Buckets:        map[string]uint64{},
		}
		m.Names[key] = ms
	}
	ms.Counters[labelValue] += delta
}

func (m *Metrics) Add(name string, labelName string, labelValue string, duration float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	key := name + "|" + labelName
	ms, exists := m.Names[key]
	if !exists {
		ms = MetricSet{
			Counters:       map[string]uint64{},
			DurationCounts: map[string]uint64{},
			DurationSums:   map[string]float64{},
			Buckets:        map[string]uint64{},
		}
		m.Names[key] = ms
	}
	ms.DurationCounts[labelValue]++
	ms.DurationSums[labelValue] += duration
	for i := len(m.Buckets) - 1; i >= 0; i-- {
		b := m.Buckets[i]
		if b.Value < duration {
			break
		}
		ms.Buckets[b.Name]++
	}
}

func (m *Metrics) Write(writer *http.ResponseWriter) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	(*writer).Header().Set("Content-Encoding", "gzip")
	gw := gzip.NewWriter((*writer))
	defer gw.Close()
	names := make([]string, 0, len(m.Names))
	for name := range m.Names {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		ms := m.Names[name]
		parts := strings.SplitN(name, "|", 2)
		metricName := parts[0]
		labelName := parts[1]
		// counters
		if len(ms.Counters) > 0 {
			gw.Write([]byte("# TYPE " + metricName + " counter\n"))
			gw.Write([]byte("# HELP " + metricName + " A counter of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
			keys := make([]string, 0, len(ms.Counters))
			for key := range ms.Counters {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, k := range keys {
				c := ms.Counters[k]
				gw.Write([]byte(metricName + "_total{" + labelName + "=" + strconv.Quote(k) + "} " + strconv.FormatUint(c, 10) + "\n"))
			}
		}
		// DurationCounts
		if len(ms.DurationCounts) > 0 {
			gw.Write([]byte("# TYPE " + metricName + "_seconds summary\n"))
			gw.Write([]byte("# UNIT " + metricName + "_seconds seconds\n"))
			gw.Write([]byte("# HELP " + metricName + "_seconds A summary of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
			keys := make([]string, 0, len(ms.DurationCounts))
			for key := range ms.DurationCounts {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			count := uint64(0)
			sum := float64(0)
			for _, k := range keys {
				c := ms.DurationCounts[k]
				count += c
				gw.Write([]byte(metricName + "_seconds_count{" + labelName + "=" + strconv.Quote(k) + "} " + strconv.FormatUint(c, 10) + "\n"))
				s := ms.DurationSums[k]
				sum += s
				gw.Write([]byte(metricName + "_seconds_sum{" + labelName + "=" + strconv.Quote(k) + "} " + strconv.FormatFloat(s, 'f', 3, 64) + "\n"))
			}
			// totals
			gw.Write([]byte("# TYPE " + metricName + "_total_seconds histogram\n"))
			gw.Write([]byte("# UNIT " + metricName + "_total_seconds seconds\n"))
			gw.Write([]byte("# HELP " + metricName + "_total_seconds A histogram of the " + strings.ReplaceAll(metricName, "_", " ") + ".\n"))
			for _, b := range m.Buckets {
				v := ms.Buckets[b.Name]
				gw.Write([]byte(metricName + "_total_seconds_bucket{le=" + strconv.Quote(b.Name) + "} " + strconv.FormatUint(v, 10) + "\n"))
			}
			gw.Write([]byte(metricName + "_total_seconds_bucket{le=\"+Inf\"} " + strconv.FormatUint(count, 10) + "\n"))
			gw.Write([]byte(metricName + "_total_seconds_sum " + strconv.FormatFloat(sum, 'f', 3, 64) + "\n"))
			gw.Write([]byte(metricName + "_total_seconds_count " + strconv.FormatUint(count, 10) + "\n"))
		}
	}
	gw.Write([]byte("# EOF\n"))
}

func (m *Metrics) AddMetrics(s2 *Metrics) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for key, value := range s2.Names {
		ms, exists := m.Names[key]
		if !exists {
			m.Names[key] = value
		} else {
			for k, v := range value.Counters {
				ms.Counters[k] += v
			}
			for k, v := range value.DurationCounts {
				ms.DurationCounts[k] += v
			}
			for k, v := range value.DurationSums {
				ms.DurationSums[k] += v
			}
			for k, v := range value.Buckets {
				ms.Buckets[k] += v
			}
		}
	}
}

func (m *Metrics) WriteGob(writer *http.ResponseWriter) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return gob.NewEncoder((*writer)).Encode(m)
}

func (m *Metrics) ReadGob(resp *http.Response) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return gob.NewDecoder(resp.Body).Decode(m)
}
