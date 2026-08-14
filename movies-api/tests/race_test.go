package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func TestConcurrentMovieUpdate(t *testing.T) {
	baseURL := "http://localhost:8081"
	movieID := 1
	url := fmt.Sprintf("%s/api/movie/%d", baseURL, movieID)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		body, _ := json.Marshal(map[string]interface{}{
			"title": "Updated By Request A",
		})
		req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
		if err != nil {
			t.Errorf("request A: failed to build request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("request A: failed to send: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("request A: expected 200, got %d", resp.StatusCode)
		}
	}()

	go func() {
		defer wg.Done()
		body, _ := json.Marshal(map[string]interface{}{
			"duration": 999,
		})
		req, err := http.NewRequest("PATCH", url, bytes.NewReader(body))
		if err != nil {
			t.Errorf("request B: failed to build request: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("request B: failed to send: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("request B: expected 200, got %d", resp.StatusCode)
		}
	}()

	wg.Wait()

	// Both requests reported success — but did both changes actually stick?
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("failed to GET movie after updates: %v", err)
	}
	defer resp.Body.Close()

	var movie struct {
		Title    string  `json:"title"`
		Duration float64 `json:"duration"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&movie); err != nil {
		t.Fatalf("failed to decode movie response: %v", err)
	}

	if movie.Title != "Updated By Request A" {
		t.Errorf("expected title to be updated to 'Updated By Request A', got: %q — request A's change was lost", movie.Title)
	}
	if movie.Duration != 999 {
		t.Errorf("expected duration to be updated to 999, got: %v — request B's change was lost", movie.Duration)
	}
}
