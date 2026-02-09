package discovery

import (
	"net"

	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
)

// Membership represents a wrapper around serf to provide discovery
type Membership struct {
	Config
	handler Handler
	serf    *serf.Serf
	events  chan serf.Event
	logger  *zap.Logger
}

// Config represents configuration parameter for nodes
type Config struct {
	NodeName       string            // NodeName
	BindAddr       string            // BindAddr
	Tags           map[string]string // Tags
	StartJoinAddrs []string          // StartJoinAddrs
}

// Handler represents an interface that member should implement
type Handler interface {
	Join(name, addr string) error
	Leave(name string) error
}

// New create a new Membership
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

// setupSerf sets up serf configuration
func (m *Membership) setupSerf() (err error) {
	addr, err := net.ResolveTCPAddr("tcp", m.BindAddr)
	if err != nil {
		return err
	}

	config := serf.DefaultConfig()
	config.Init()
	config.MemberlistConfig.BindAddr = addr.IP.String()
	config.MemberlistConfig.BindPort = addr.Port

	m.events = make(chan serf.Event)
	config.EventCh = m.events

	config.Tags = m.Tags
	config.NodeName = m.Config.NodeName
	m.serf, err = serf.Create(config)
	if err != nil {
		return err
	}

	go m.evenHandler()

	if m.StartJoinAddrs != nil {
		_, err := m.serf.Join(m.StartJoinAddrs, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// evenHandler handles different event type shuch as EventMemberJoin, EventMemberLeave
func (m *Membership) evenHandler() {
	for e := range m.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}

				m.handleJoin(member)
			}

		case serf.EventMemberLeave, serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					return
				}

				m.handleLeave(member)
			}
		}
	}
}

// handleJoin handles join event member
func (m *Membership) handleJoin(member serf.Member) {
	err := m.handler.Join(member.Name, member.Tags["rpc_addr"])
	if err != nil {
		m.logError(err, "failed to join", member)
	}
}

// handleLeave handles leave event member
func (m *Membership) handleLeave(member serf.Member) {
	if err := m.handler.Leave(member.Name); err != nil {
		m.logError(err, "failed to leave", member)
	}
}

// isLocal checks whether a given member is the local member(it self)
func (m *Membership) isLocal(member serf.Member) bool {
	return m.serf.LocalMember().Name == member.Name
}

// Members returns members(node) of the cluster
func (m *Membership) Members() []serf.Member {
	return m.serf.Members()
}

// logError logs error
func (m *Membership) logError(err error, msg string, member serf.Member) {
	m.logger.Error(
		msg,
		zap.Error(err),
		zap.String("name", member.Name),
		zap.String("rpc_addr", member.Tags["rpc_addr"]),
	)
}

// Leave tells a member to leave the cluster
func (m *Membership) Leave() error {
	return m.serf.Leave()
}
