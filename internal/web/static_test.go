package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestStaticFileHandler_ServesJavaScript verifies sprite.js is embedded and servable
func TestStaticFileHandler_ServesJavaScript(t *testing.T) {
	handler := StaticFileHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	// Request sprite.js
	resp, err := http.Get(server.URL + "/js/sprite.js")
	if err != nil {
		t.Fatalf("Failed to get sprite.js: %v", err)
	}
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "javascript") && !strings.Contains(contentType, "text/plain") {
		t.Errorf("Expected JavaScript content type, got %s", contentType)
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	bodyStr := string(body)

	// Verify key components exist in sprite.js
	expectedStrings := []string{
		"class Sprite",
		"class TextureAtlas",
		"class SpriteRenderer",
		"function drawSprite",
		"function loadImage",
		"function loadAtlas",
		"zIndex",
		"rotation",
		"texture",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(bodyStr, expected) {
			t.Errorf("Expected sprite.js to contain '%s'", expected)
		}
	}
}

// TestStaticFileHandler_404ForMissing verifies 404 for missing files
func TestStaticFileHandler_404ForMissing(t *testing.T) {
	handler := StaticFileHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/js/nonexistent.js")
	if err != nil {
		t.Fatalf("Failed to get nonexistent file: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for missing file, got %d", resp.StatusCode)
	}
}

// TestStaticFileHandler_SecurityNoDirectoryTraversal verifies path traversal protection
func TestStaticFileHandler_SecurityNoDirectoryTraversal(t *testing.T) {
	handler := StaticFileHandler()
	server := httptest.NewServer(handler)
	defer server.Close()

	// Attempt directory traversal
	resp, err := http.Get(server.URL + "/../../../etc/passwd")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Should get 404 or error, not success
	if resp.StatusCode == http.StatusOK {
		t.Error("Directory traversal should not succeed")
	}
}
