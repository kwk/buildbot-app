package main

import (
	"fmt"
	"log"
	"net/http"
)

func (srv *AppServer) HandleGithubHook() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Println("/github-hook")
		err := srv.GithubEventHandler.HandleEventRequest(req)
		if err != nil {
			log.Printf("error while processing request: %+v", err)
			fmt.Println("error")
		}
	}
}
