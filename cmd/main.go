package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	api "github.com/nico-phil/go-log/api/v1"
	llog "github.com/nico-phil/go-log/internal/log"
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

	rec1 := api.Record{
		Value: []byte("hello world"),
	}

	off, err := wLog.Append(&rec1)
	if err != nil {
		log.Fatal("append:", err)
	}

	fmt.Println("off:", off)
	readRec, err := wLog.Read(off)
	if err != nil {
		log.Fatal("read Record", err)
	}

	fmt.Println("readRec:", readRec)

	wLog.Close()

}
