package main

import (
	"fmt"
	"os"

	"github.com/nico-phil/go-log/internal/log"
)

func main() {
	f, err := os.Create("storelog")
	if err != nil {
		return
	}

	s, _ := log.NewStore(f)

	n, pos, err := s.Append([]byte("first"))
	if err != nil {
		return
	}

	fmt.Printf("n: %d\n", n)
	fmt.Printf("pos: %d\n", pos)

	d1, _ := s.Read(pos)
	fmt.Println("d1", string(d1))

	n, pos, err = s.Append([]byte("second"))
	if err != nil {
		return
	}

	fmt.Printf("n: %d\n", n)
	fmt.Printf("pos: %d\n", pos)

	d2, _ := s.Read(pos)
	fmt.Println("d2", string(d2))
	fmt.Println("------------------------------------")

	fIn, err := os.Create("index_test")
	if err != nil {
		return
	}
	c := log.Config{}
	c.Segment.MaxIndexBytes = 1024
	newIndexFile, err := log.NewIndex(fIn, c)
	if err != nil {
		fmt.Println("newIndexFile:", err)
		return
	}

	newIndexFile.Write(1, 10)
	newIndexFile.Write(2, 6)
	resBuf := make([]byte, 1024)

	fmt.Println(newIndexFile.MMap)

	nu, _ := f.Read(resBuf)
	fmt.Println("resBuf", string(resBuf[0:nu]))

}
