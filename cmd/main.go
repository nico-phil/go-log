package main

import (
	"fmt"
	"log"

	api "github.com/nico-phil/go-log/api/v1"
	llog "github.com/nico-phil/go-log/internal/log"
)

func main() {
	c := llog.Config{}
	c.Segment.MaxIndexBytes = 1024
	sg, err := llog.NewSegment("data", 0, c)
	if err != nil {

		log.Fatal(err)
	}

	log.Printf("%+v", *sg)
	r := api.Record{
		Value: []byte("hello world"),
	}

	off, err := sg.Append(&r)
	if err != nil {
		log.Fatal("append:", err)
	}

	fmt.Println(off)

}
