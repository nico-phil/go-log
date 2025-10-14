package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type HttpServer struct {
	log *Log
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		log: NewLog(),
	}
}

type ProduceRequest struct {
	Record Record `json:"record"`
}

type ProduceResponse struct {
	Offset int `json:"offset"`
}

type ConsumeRequest struct {
	Offset uint64
}

type ConsumeResponse struct {
}

func (s *HttpServer) handleProduce(w http.ResponseWriter, r *http.Request) {
	var req ProduceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	offset, err := s.log.Append(req.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pr := ProduceResponse{Offset: offset}

	err = json.NewEncoder(w).Encode(pr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *HttpServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	var req ConsumeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := s.log.Read(req.Offset)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	err = json.NewEncoder(w).Encode(&record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func New(addr int) *http.Server {
	httpServer := NewHttpServer()
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", httpServer.handleProduce)
	mux.HandleFunc("GET /", httpServer.handleConsume)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", addr),
		Handler: mux,
	}
}
