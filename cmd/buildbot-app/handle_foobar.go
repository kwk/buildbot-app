package main

import (
	"fmt"
	"log"
	"net/http"
)

func (srv *AppServer) HandleFoobar() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("/foobar called")
		fmt.Fprintf(w, "pong")
	}
}
