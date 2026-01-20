package discovery

import (
	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
)

type Membership struct {
	Config
	handler Handler
	serf    *serf.Serf
	events  chan serf.Serf
	logger  *zap.Logger
}

type Config struct {
	NodeName       string
	BindAddress    string
	Tags           map[string]string
	StartJoinAddrs []string
}

type Handler interface {
	Join(name, addr string) error
	Leave(name string) error
}

func New(handler Handler, config Config) (*Membership, error) {
	c := Membership{
		Config:  config,
		handler: handler,
		logger:  zap.L().Named("membership"),
	}
	err := c.setupSerf()
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (m *Membership) setupSerf() (err error) {
	return nil
}
