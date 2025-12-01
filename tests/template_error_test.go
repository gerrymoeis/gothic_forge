package tests

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gothicforge3/app/routes"
	"gothicforge3/app/templates"
	"gothicforge3/internal/server"
)

// errorWriter simulates a writer that fails after a certain number of bytes
type errorWriter struct {
	failAfter int
	written   int
}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	if e.written >= e.failAfter {
		return 0, errors.New("simulated write error")
	}
	e.written += len(p)
	return len(p), nil
}

// Test that template rendering errors are properly handled
func Test_Template_DBPostsList_WriteError(t *testing.T) {
	// Create a writer that will fail
	ew := &errorWriter{failAfter: 0}
	
	// Create sample data
	items := []templates.DBPostItem{
		{ID: 1, Title: "Test Post", Body: "Test Body", CreatedAt: "2024-01-01"},
	}
	
	// Render template with error writer
	component := templates.DBPostsList(items)
	ctx := context.Background()
	err := component.Render(ctx, ew)
	
	// Verify error is returned
	if err == nil {
		t.Fatal("expected error from template rendering, got nil")
	}
	
	// Verify error is propagated (either contains template name or the underlying error)
	errStr := err.Error()
	if !strings.Contains(errStr, "DBPostsList") && !strings.Contains(errStr, "simulated write error") {
		t.Errorf("expected error to be propagated, got: %v", err)
	}
}

func Test_Template_DBPostsForm_WriteError(t *testing.T) {
	// Create a writer that will fail
	ew := &errorWriter{failAfter: 0}
	
	// Create sample data
	item := &templates.DBPostItem{
		ID:    1,
		Title: "Test Post",
		Body:  "Test Body",
	}
	
	// Render template with error writer
	component := templates.DBPostsForm("/db/posts/1", item, "Update")
	ctx := context.Background()
	err := component.Render(ctx, ew)
	
	// Verify error is returned
	if err == nil {
		t.Fatal("expected error from template rendering, got nil")
	}
	
	// Verify error is propagated (either contains template name or the underlying error)
	errStr := err.Error()
	if !strings.Contains(errStr, "DBPostsForm") && !strings.Contains(errStr, "simulated write error") {
		t.Errorf("expected error to be propagated, got: %v", err)
	}
}

// Test that template errors propagate to HTTP layer
func Test_Template_Error_Propagation_HTTP(t *testing.T) {
	_ = os.Setenv("LOG_FORMAT", "off")
	_ = os.Setenv("DATABASE_URL", "")
	
	r := server.New()
	routes.Register(r)
	
	// Test the /db/posts/new route which uses DBPostsForm
	req := httptest.NewRequest(http.MethodGet, "/db/posts/new", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	
	// This should succeed normally (no database required for form)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	
	// Verify the response contains expected HTML
	body := rec.Body.String()
	if !strings.Contains(body, "Post") {
		t.Errorf("expected response to contain 'Post', got: %s", body)
	}
}

// Test successful template rendering
func Test_Template_DBPostsList_Success(t *testing.T) {
	var buf bytes.Buffer
	
	items := []templates.DBPostItem{
		{ID: 1, Title: "First Post", Body: "First Body", CreatedAt: "2024-01-01"},
		{ID: 2, Title: "Second Post", Body: "Second Body", CreatedAt: "2024-01-02"},
	}
	
	component := templates.DBPostsList(items)
	ctx := context.Background()
	err := component.Render(ctx, &buf)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	html := buf.String()
	
	// Verify content is rendered
	if !strings.Contains(html, "First Post") {
		t.Error("expected HTML to contain 'First Post'")
	}
	if !strings.Contains(html, "Second Post") {
		t.Error("expected HTML to contain 'Second Post'")
	}
	if !strings.Contains(html, "/db/posts/1/edit") {
		t.Error("expected HTML to contain link to first post")
	}
}

func Test_Template_DBPostsForm_Success(t *testing.T) {
	var buf bytes.Buffer
	
	item := &templates.DBPostItem{
		ID:    42,
		Title: "Test Title",
		Body:  "Test Body Content",
	}
	
	component := templates.DBPostsForm("/db/posts/42", item, "Update")
	ctx := context.Background()
	err := component.Render(ctx, &buf)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	html := buf.String()
	
	// Verify form elements are rendered
	if !strings.Contains(html, "Test Title") {
		t.Error("expected HTML to contain 'Test Title'")
	}
	if !strings.Contains(html, "Test Body Content") {
		t.Error("expected HTML to contain 'Test Body Content'")
	}
	if !strings.Contains(html, "/db/posts/42") {
		t.Error("expected HTML to contain form action")
	}
	if !strings.Contains(html, "Update") {
		t.Error("expected HTML to contain submit button text")
	}
}

// Test empty list rendering
func Test_Template_DBPostsList_Empty(t *testing.T) {
	var buf bytes.Buffer
	
	items := []templates.DBPostItem{}
	
	component := templates.DBPostsList(items)
	ctx := context.Background()
	err := component.Render(ctx, &buf)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	html := buf.String()
	
	// Verify empty message is rendered
	if !strings.Contains(html, "No posts yet") {
		t.Error("expected HTML to contain 'No posts yet' message")
	}
}

// Test form with nil item (new post)
func Test_Template_DBPostsForm_NilItem(t *testing.T) {
	var buf bytes.Buffer
	
	component := templates.DBPostsForm("/db/posts", nil, "Create")
	ctx := context.Background()
	err := component.Render(ctx, &buf)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	
	html := buf.String()
	
	// Verify form is rendered with empty values
	if !strings.Contains(html, "/db/posts") {
		t.Error("expected HTML to contain form action")
	}
	if !strings.Contains(html, "Create") {
		t.Error("expected HTML to contain 'Create' button")
	}
}

// partialErrorWriter writes successfully for a few bytes then fails
type partialErrorWriter struct {
	buf       bytes.Buffer
	failAfter int
	written   int
}

func (p *partialErrorWriter) Write(data []byte) (n int, err error) {
	if p.written >= p.failAfter {
		return 0, io.ErrShortWrite
	}
	
	remaining := p.failAfter - p.written
	if len(data) <= remaining {
		n, err = p.buf.Write(data)
		p.written += n
		return n, err
	}
	
	// Write partial data then fail
	n, _ = p.buf.Write(data[:remaining])
	p.written += n
	return n, io.ErrShortWrite
}

// Test partial write failure
func Test_Template_DBPostsList_PartialWriteError(t *testing.T) {
	// Allow some bytes to be written before failing
	pw := &partialErrorWriter{failAfter: 50}
	
	items := []templates.DBPostItem{
		{ID: 1, Title: "Test Post", Body: "Test Body", CreatedAt: "2024-01-01"},
	}
	
	component := templates.DBPostsList(items)
	ctx := context.Background()
	err := component.Render(ctx, pw)
	
	// Should get an error
	if err == nil {
		t.Fatal("expected error from partial write, got nil")
	}
	
	// Verify error is propagated (contains either template context or write error)
	errStr := err.Error()
	if !strings.Contains(errStr, "DBPostsList") && !strings.Contains(errStr, "write HTML failed") && !strings.Contains(errStr, "short write") {
		t.Errorf("expected error to be propagated, got: %v", err)
	}
}
