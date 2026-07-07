package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	api "github.com/nico-phil/go-log/api/v1"
	"github.com/nico-phil/go-log/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	addr := flag.String("addr", ":8400", "service addr")

	peerTlsConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile: config.RootClientCertFile,
		KeyFile:  config.RootClientKeyFile,
		CAFile:   config.CAFile,
		Server:   false,
	})

	tlsCreds := credentials.NewTLS(peerTlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}

	conn, err := grpc.NewClient(*addr, opts...)
	if err != nil {
		return
	}

	client := api.NewLogClient(conn)

	r, err := client.GetServers(context.Background(), &api.GetServersRequest{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("servers:")
	for _, s := range r.Servers {
		fmt.Printf("\t- %v\n", s)
	}

}
