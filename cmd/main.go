package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	llog "github.com/nico-phil/go-log/internal/log"
	"github.com/nico-phil/go-log/internal/server"
)

func main() {
	err := os.Mkdir("data", os.ModeDir)
	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Fatal("error creating dir:", err)
	}
	c := llog.Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = 32

	wLog, err := llog.NewLog("data", c)
	if err != nil {
		log.Fatal("error-Newlog:", err)
	}

	config := server.Config{
		CommitLog: wLog,
	}

	srv, err := server.NewGRPCServer(&config)
	if err != nil {
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 4000))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("lis: %+v", lis)
	err = srv.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Hello")
}
