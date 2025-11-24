package main

import (
	"fmt"
	"log"
	"os"

	llog "github.com/nico-phil/go-log/internal/log"
)

func main() {
	f, err := os.Create("store-test")
	if err != nil {
		log.Fatal("create file: ", err)
	}
	s, err := llog.NewStore(f)
	if err != nil {
		log.Fatal("create store:")
	}

	fIndex, err := os.Create("index-test")
	if err != nil {
		log.Fatal("index file:", err)
	}
	c := llog.Config{}
	c.Segment.MaxIndexBytes = 1024
	index, err := llog.NewIndex(fIndex, c)
	if err != nil {
		log.Fatal("newIndex:", err)
	}

	for i := 0; i < 10; i++ {
		v := fmt.Sprintf("hello_%d", i)
		_, pos, _ := s.Append([]byte(v))
		// fmt.Printf("offset: %d value: %s\n", pos, v)

		s.Read(pos)
		// fmt.Printf("read value: %s\n", string(r))
		index.Write(uint32(i), pos)

		out, pos, _ := index.Read(int64(i))
		fmt.Printf("index- out:%d  pos: %d\n", out, pos)

	}

	r, _ := s.Read(15)
	fmt.Printf("read value: %s\n", string(r))

}
