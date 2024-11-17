package metrics

import (
	"compress/gzip"
	"io"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// TestAdd adds a value and checks that the value is stored.
func TestAdd(t *testing.T) {
	stats := New()
	stats.Add("name", "label", "value", 1.23)
	w := httptest.NewRecorder()
	stats.Write(w)
	resp := w.Result()
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Errorf("error reading gz: %q", err.Error())
	}
	body, err := io.ReadAll(gz)
	if err != nil {
		t.Errorf("error reading body: %q", err.Error())
	}
	got := string(body)
	want := "name_seconds_count{label=\"value\"} 1\nname_seconds_sum{label=\"value\"} 1.230\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}

// TestInc increments a value and checks that the value is stored.
func TestInc(t *testing.T) {
	stats := New()
	stats.Inc("name", "label", "value", 1)
	w := httptest.NewRecorder()
	stats.Write(w)
	resp := w.Result()
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Errorf("error reading gz: %q", err.Error())
	}
	body, err := io.ReadAll(gz)
	if err != nil {
		t.Errorf("error reading body: %q", err.Error())
	}
	got := string(body)
	want := "name_total{label=\"value\"} 1\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}

// TestAddMetrics adds two metrics and checks that the sum is stored.
func TestAddMetrics(t *testing.T) {
	stats := New()
	stats.Add("name", "label", "value", 1.23)
	stats2 := New()
	stats2.Add("name", "label", "value", 1.23)
	stats.AddMetrics(stats2)
	w := httptest.NewRecorder()
	stats.Write(w)
	resp := w.Result()
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Errorf("error reading gz: %q", err.Error())
	}
	body, err := io.ReadAll(gz)
	if err != nil {
		t.Errorf("error reading body: %q", err.Error())
	}
	got := string(body)
	want := "name_seconds_count{label=\"value\"} 2\nname_seconds_sum{label=\"value\"} 2.460\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}

// TestAddParallel adds a value 10k times and checks that the value is stored.
func TestAddParallel(t *testing.T) {
	stats := New()
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats.Add("name", "label", "value", 1.23)
		}()
	}
	wg.Wait()
	w := httptest.NewRecorder()
	stats.Write(w)
	resp := w.Result()
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Errorf("error reading gz: %q", err.Error())
	}
	body, err := io.ReadAll(gz)
	if err != nil {
		t.Errorf("error reading body: %q", err.Error())
	}
	got := string(body)
	want := "name_seconds_count{label=\"value\"} 10000\nname_seconds_sum{label=\"value\"} 12300.000\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}

// TestIncParallel increments a value 10k times and checks that the value is stored.
func TestIncParallel(t *testing.T) {
	stats := New()
	var wg sync.WaitGroup
	for i := 0; i < 10000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			stats.Inc("name", "label", "value", 1)
		}()
	}
	wg.Wait()
	w := httptest.NewRecorder()
	stats.Write(w)
	resp := w.Result()
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Errorf("error reading gz: %q", err.Error())
	}
	body, err := io.ReadAll(gz)
	if err != nil {
		t.Errorf("error reading body: %q", err.Error())
	}
	got := string(body)
	want := "name_total{label=\"value\"} 10000\n"
	if !strings.Contains(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}
