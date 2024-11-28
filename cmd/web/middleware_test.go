package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"judaicaswap.com/internal/assert"
)

func TestCommonHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	commonHeaders(next).ServeHTTP(rr, r)

	rs := rr.Result()

	expectedValue := "default-src 'self'; img-src 'self' https://judaicaswap.s3.us-east-1.amazonaws.com/; " +
		"font-src fonts.googleapis.com; style-src 'self'; require-trusted-types-for 'script'"
	assert.Equal(t, rs.Header.Get("Content-Security-Policy"), expectedValue)

	// Check that the middleware has correctly set the Referrer-Policy
	// header on the response.
	expectedValue = "strict-origin-when-cross-origin"
	assert.Equal(t, rs.Header.Get("Referrer-Policy"), expectedValue)

	// Check that the middleware has correctly set the X-Content-Type-Options
	// header on the response.
	expectedValue = "nosniff"
	assert.Equal(t, rs.Header.Get("X-Content-Type-Options"), expectedValue)

	// Check that the middleware has correctly set the X-Frame-Options header
	// on the response.
	expectedValue = "deny"
	assert.Equal(t, rs.Header.Get("X-Frame-Options"), expectedValue)

	// Check that the middleware has correctly set the X-XSS-Protection header
	// on the response
	expectedValue = "0"
	assert.Equal(t, rs.Header.Get("X-XSS-Protection"), expectedValue)

	// Check that the middleware has correctly called the next handler in line
	// and the response status code and body are as expected.
	assert.Equal(t, rs.StatusCode, http.StatusOK)

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	assert.Equal(t, string(body), "OK")

}
