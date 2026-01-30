package main

import (
	"fmt"
	"os"

	llog "github.com/nico-phil/go-log/internal/log"
)

func main() {
	f, err := os.OpenFile("index-demo", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("OpenFile: %v\n", err)
		return
	}
	conf := llog.Config{}
	conf.Segment.MaxIndexBytes = 1024

	idx, err := llog.NewIndex(f, conf)
	if err != nil {
		fmt.Printf("NewIndex: %v\n", err)
		return
	}
	defer idx.Close()

	err = idx.Write(0, 10)
	if err != nil {
		fmt.Printf("WRITE: %v\n", err)
	}

	out, pos, err := idx.Read(0)
	if err != nil {
		fmt.Printf("READ: %v\n", err)
		return
	}

	fmt.Println("out", out)
	fmt.Println("pos", pos)

	err = idx.Write(1, 18)
	if err != nil {
		fmt.Printf("WRITE: %v\n", err)
	}
	out, pos, err = idx.Read(1)
	if err != nil {
		fmt.Printf("READ: %v\n", err)
		return
	}

	fmt.Println("out", out)
	fmt.Println("pos", pos)

	out, pos, err = idx.Read(-1)
	if err != nil {
		fmt.Printf("READ: %v\n", err)
		return
	}
	fmt.Println("last-out", out)
	fmt.Println("last-pos", pos)

}
