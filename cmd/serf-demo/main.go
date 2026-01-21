package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hashicorp/serf/serf"
)

// Config represents configuration for the cluster
type Config struct {
	// The node name as the nodes's unique identifier across the serf cluster. if the node name is not provided, serf uses the hostname
	NodeName string
	// BindAddr and BindPort: serf listens on this adderss and port for gossiping
	BindAddr       string
	Tags           map[string]string
	StartJoinAddrs []string
}

func getEnvVariable(envV string) string {
	return os.Getenv(envV)
}

func main() {

	nodeName := getEnvVariable("NODE_NAME")
	nodeAddr := getEnvVariable("BIND_ADDR")
	port := getEnvVariable("SERF_PORT")

	joinAddr := getEnvVariable("JOIN_DDR")

	portInt, _ := strconv.Atoi(port)

	// serf provide 2 methods used to initialize a configration,

	conf := serf.DefaultConfig()
	conf.Init()

	conf.NodeName = nodeName
	conf.MemberlistConfig.BindAddr = nodeAddr
	conf.MemberlistConfig.BindPort = portInt

	events := make(chan serf.Event)
	conf.EventCh = events

	instance, err := serf.Create(conf)
	if err != nil {
		panic(err.Error())
	}

	go func() {
		for e := range events {
			switch ev := e.(type) {
			case serf.MemberEvent:
				for _, m := range ev.Members {
					switch ev.EventType() {
					case serf.EventMemberJoin:
						fmt.Println("[JOIN]: Node joinded:", m.Name, m.Addr)
					case serf.EventMemberLeave:
						fmt.Println("[LEAVE]: Node leave gracefully:", m.Name, m.Addr)
					case serf.EventMemberFailed:
						fmt.Println("[FAILED]: Node failed:", m.Name, m.Addr)
					default:
						fmt.Println("[OTHER]: event:", ev.EventType(), m.Name)
					}

				}
			}
		}
	}()

	if joinAddr != "" {
		instance.Join([]string{joinAddr}, false)
	}

	select {}
}
