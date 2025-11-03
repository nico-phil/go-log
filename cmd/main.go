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

}
