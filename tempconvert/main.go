package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"example.com/user/tempconvert/contract"
	"google.golang.org/protobuf/proto"
)

// GetMux returns the multiplexer - registered routes & functions
func GetMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)
	mux.HandleFunc("/celsius", celsiusHandler)
	return mux
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!"))
}

func celsiusHandler(w http.ResponseWriter, r *http.Request) {
	fahrenheit, ok := r.URL.Query()["fahrenheit"]
	if !ok {
		http.Error(w, "missing fahrenheit URL query param", http.StatusBadRequest)
		return
	}
	f, err := strconv.ParseFloat(fahrenheit[0], 64)
	if err != nil {
		http.Error(w, "invalid fahrenheit value", http.StatusBadRequest)
		return
	}
	reply := celsiusResolver(&contract.TempConvertRequest{Fahrenheit: f})
	body, err := celsiusMarshaller(w, r, reply)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(body)
}

func celsiusResolver(r *contract.TempConvertRequest) *contract.TempConvertReply {
	c := (r.Fahrenheit - 32) * 5 / 9
	return &contract.TempConvertReply{Celsius: c}
}

func celsiusMarshaller(w http.ResponseWriter, r *http.Request, reply *contract.TempConvertReply) ([]byte, error) {
	accept := r.Header.Get("accept")
	w.Header().Set("Content-Type", accept)
	switch accept {
	case "application/protobuf":
		return proto.Marshal(reply)
	case "application/json":
		return json.Marshal(reply)
	default:
		w.Header().Set("Content-Type", "text/plain")
		return []byte(strconv.FormatFloat(reply.Celsius, 'g', -1, 64)), nil
	}
}

func main() {
	s := &http.Server{
		Handler:      GetMux(),
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  1 * time.Minute,
	}
	log.Fatal(s.ListenAndServe())
}
