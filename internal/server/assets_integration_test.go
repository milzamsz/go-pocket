package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssetsRoute_ServesEmbeddedFallback(t *testing.T) {
	t.Parallel()

	_, mux := newRouteTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/assets/js/app.min.js", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.NotEmpty(t, rec.Body.String())
	require.Contains(t, rec.Header().Get("Content-Type"), "javascript")
}

func TestAssetsRoute_ServesLocalFileFirst(t *testing.T) {
	t.Parallel()

	_, mux := newRouteTestServer(t)
	localPath := filepath.FromSlash("assets/css/test-local.css")
	require.NoError(t, os.MkdirAll(filepath.Dir(localPath), 0o755))
	require.NoError(t, os.WriteFile(localPath, []byte("/* local css */"), 0o644))
	t.Cleanup(func() { _ = os.Remove(localPath) })

	status, body := performRequest(t, mux, http.MethodGet, "/assets/css/test-local.css", "")
	require.Equal(t, http.StatusOK, status)
	require.Contains(t, body, "local css")
}

func TestAssetsRoute_MissingAssetReturnsNotFound(t *testing.T) {
	t.Parallel()

	_, mux := newRouteTestServer(t)
	status, _ := performRequest(t, mux, http.MethodGet, "/assets/css/does-not-exist.css", "")
	require.Equal(t, http.StatusNotFound, status)
}

func TestAssetsRoute_OutputCSSReachableWithCSSType(t *testing.T) {
	t.Parallel()

	_, mux := newRouteTestServer(t)
	outputPath := filepath.FromSlash("assets/css/output.css")

	existing, readErr := os.ReadFile(outputPath)
	if readErr == nil {
		t.Cleanup(func() { _ = os.WriteFile(outputPath, existing, 0o644) })
	} else {
		require.NoError(t, os.MkdirAll(filepath.Dir(outputPath), 0o755))
		require.NoError(t, os.WriteFile(outputPath, []byte("body{margin:0}"), 0o644))
		t.Cleanup(func() { _ = os.Remove(outputPath) })
	}

	req := httptest.NewRequest(http.MethodGet, "/assets/css/output.css", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Header().Get("Content-Type"), "text/css")
	require.NotEmpty(t, rec.Body.String())
}

func TestAssetsRoute_LocalSortableJSReachable(t *testing.T) {
	t.Parallel()

	_, mux := newRouteTestServer(t)
	status, body := performRequest(t, mux, http.MethodGet, "/assets/js/sortable.min.js", "")
	require.Equal(t, http.StatusOK, status)
	require.Contains(t, body, "Sortable")
}
