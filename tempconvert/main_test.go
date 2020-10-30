package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var mux http.Handler

func mockRequest(method string, url string) ([]byte, *http.Response, error) {
	request := httptest.NewRequest(method, url, nil)
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, request)
	resp := recorder.Result()
	body, err := ioutil.ReadAll(recorder.Body)
	return body, resp, err
}

// This is a special function used to run code before and after testing runs
func TestMain(m *testing.M) {
	// Code here runs before testing starts
	mux = GetMux()
	// Run tests
	exitCode := m.Run()
	// Code here runs after testing finishes
	os.Exit(exitCode)
}

func TestHelloOk(t *testing.T) {
	body, resp, err := mockRequest("GET", "http://localhost:1234/")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	require.Equal(t, "Hello World!", string(body))
}

func TestConvertOk(t *testing.T) {
	body, resp, err := mockRequest("GET", "http://localhost:1234/celsius?fahrenheit=32")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, string(body))
	require.Equal(t, "0", string(body))
}
