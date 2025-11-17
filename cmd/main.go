package main

import (
	"fmt"
	"log"

	api "github.com/nico-phil/go-log/api/v1"

	llog "github.com/nico-phil/go-log/internal/log"
)

func main() {
	c := llog.Config{}
	c.Segment.MaxIndexBytes = 12 * 3
	c.Segment.MaxStoreBytes = 1024
	sg, err := llog.NewSegment("data", 16, c)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v", *sg)
	r := api.Record{
		Value: []byte("hello world"),
	}

	off, err := sg.Append(&r)
	if err != nil {
		log.Fatal("append(main):\n", err)
	}

	fmt.Println(off)

}
