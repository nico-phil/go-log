package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	llog "github.com/nico-phil/go-log/internal/log"
)

func main() {
	err := os.Mkdir("data", os.ModeDir)
	if err != nil && !errors.Is(err, os.ErrExist) {
		log.Fatal("error creating dir:", err)
	}
	c := llog.Config{}

	wLog, err := llog.NewLog("data", c)
	if err != nil {
		log.Fatal("error-Newlog:", err)
	}

	fmt.Printf("wlog: %+v\n", wLog)

}
