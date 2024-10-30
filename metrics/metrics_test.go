package metrics

import (
	"compress/gzip"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	want := "name_seconds_count{label=\"value\"} 1\nname_seconds_sum{label=\"value\"} 1.230"
	if !strings.Contains(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}

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
	want := "name_total{label=\"value\"} 1"
	if !strings.Contains(got, want) {
		t.Errorf("got %s, want %s", got, want)
	}
}
