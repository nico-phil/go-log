package main

import (
	"fmt"
	"log"

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
	defer sg.Close()

	fmt.Printf("%+v\n", *sg)
	// r := api.Record{
	// 	Value: []byte("hello world"),
	// }

	// off, err := sg.Append(&r)
	// if err != nil {
	// 	log.Fatal("append(main):\n", err)
	// }

	// fmt.Println("offset: ", off)

	// r1, err := sg.Read(off)
	// if err != nil {
	// 	return
	// }

	// fmt.Println("record: ", r1)

	_, pos, _ := sg.ReadIndex(0)
	fmt.Println("position:", pos)

}
