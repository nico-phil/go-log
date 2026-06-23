package log

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"

	api "github.com/nico-phil/go-log/api/v1"
	"google.golang.org/protobuf/proto"
)

// DistributedLog repesent a wrapper around raft to replicate the log
type DistributedLog struct {
	config Config
	log    *Log
	raft   *raft.Raft
}

// NewDistrubutedLog create a new distributed log
func NewDistrubutedLog(dataDir string, config Config) (*DistributedLog, error) {
	l := &DistributedLog{
		config: config,
	}
	if err := l.setupLog(dataDir); err != nil {
		return nil, err
	}

	if err := l.setupRaft(dataDir); err != nil {
		return nil, err
	}

	return l, nil
}

// setupLog creates the log server, where this server will store the user's records
func (l *DistributedLog) setupLog(dataDir string) error {
	logDir := path.Join(dataDir, "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	var err error
	l.log, err = NewLog(logDir, l.config)

	return err
}

// setupRaft configures and creates the server's raft instance
func (l *DistributedLog) setupRaft(dataDir string) error {
	fsm := &fsm{log: l.log}

	logDir := path.Join(dataDir, "raft", "log")
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		return err
	}

	logConfig := l.config
	logConfig.Segment.InitialOffset = 1
	logStore, err := newLogStore(logDir, logConfig)
	if err != nil {
		return err
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft", "stable"))
	if err != nil {
		return err
	}

	retain := 1
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join(dataDir, "raft"),
		retain,
		os.Stderr,
	)
	if err != nil {
		return err
	}

	maxPool := 5
	timeout := 10 * time.Second
	transport := raft.NewNetworkTransport(
		l.config.Raft.StreamLayer,
		maxPool,
		timeout,
		os.Stderr,
	)

	config := raft.DefaultConfig()
	config.LocalID = l.config.Raft.LocalID

	if l.config.Raft.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = l.config.Raft.HeartbeatTimeout
	}

	if l.config.Raft.ElectionTimeout != 0 {
		config.ElectionTimeout = l.config.Raft.ElectionTimeout
	}

	if l.config.Raft.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = l.config.Raft.LeaderLeaseTimeout
	}

	if l.config.Raft.CommitTimeout != 0 {
		config.CommitTimeout = l.config.Raft.CommitTimeout
	}

	l.raft, err = raft.NewRaft(
		config,
		fsm,
		logStore,
		stableStore,
		snapshotStore,
		transport,
	)
	if err != nil {
		return err
	}

	hasState, err := raft.HasExistingState(
		logStore,
		stableStore,
		snapshotStore,
	)
	if err != nil {
		return err
	}
	if l.config.Raft.Bootstrap && !hasState {
		config := raft.Configuration{
			Servers: []raft.Server{{
				ID:      config.LocalID,
				Address: raft.ServerAddress(l.config.Raft.BindAddr),
			}},
		}
		err = l.raft.BootstrapCluster(config).Error()
	}
	return err
}

// Append appends a record to the log.We do not append the record directly, we tell raft to apply the command
func (l *DistributedLog) Append(record *api.Record) (uint64, error) {
	res, err := l.apply(
		AppendRequestType,
		&api.ProduceRequest{Record: record},
	)
	if err != nil {
		return 0, err
	}

	off, ok := res.(*api.ProduceResponse)
	if !ok {
		return 0, err
	}
	return uint64(off.Offset), nil

}

// apply wraps raft's API to apply the request and return their response
func (l *DistributedLog) apply(reqType RequestType, req proto.Message) (interface{}, error) {
	var buf bytes.Buffer
	_, err := buf.Write([]byte{byte(reqType)})
	if err != nil {
		return nil, err
	}
	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}
	timeout := time.Second * 10
	future := l.raft.Apply(buf.Bytes(), timeout)
	if future.Error() != nil {
		return nil, future.Error()
	}

	res := future.Response()
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res, nil
}

// Read reads the record for the offset from the server' log
func (l *DistributedLog) Read(offset uint64) (*api.Record, error) {
	return l.log.Read(offset)
}

// Discovery Integration

// Join Adds the server to the raft cluster. each server is added as a voter
func (l *DistributedLog) Join(id, addr string) error {
	configFuture := l.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}
	serverID := raft.ServerID(id)
	serverAddr := raft.ServerAddress(addr)
	for _, srv := range configFuture.Configuration().Servers {
		if srv.ID == serverID || srv.Address == serverAddr {
			if srv.ID == serverID && srv.Address == serverAddr {
				// server has already joined
				return nil
			}
			// remove the existing server
			removeFuture := l.raft.RemoveServer(serverID, 0, 0)
			if err := removeFuture.Error(); err != nil {
				return err
			}
		}
	}
	addFuture := l.raft.AddVoter(serverID, serverAddr, 0, 0)
	if err := addFuture.Error(); err != nil {
		return err
	}
	return nil
}

func (l *DistributedLog) Leave(id string) error {
	removeFuture := l.raft.RemoveServer(raft.ServerID(id), 0, 0)
	return removeFuture.Error()
}

// WaitForLeader blocks until the cluster has elected a leader or times out
func (l *DistributedLog) WaitForLeader(timeout time.Duration) error {
	timeoutc := time.After(timeout)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-timeoutc:
			return fmt.Errorf("timed out")
		case <-ticker.C:
			if l := l.raft.Leader(); l != "" {
				return nil
			}
		}
	}
}

// Close shutdown the raft instance and closes the local log
func (l *DistributedLog) Close() error {
	f := l.raft.Shutdown()
	if err := f.Error(); err != nil {
		return err
	}
	return l.log.Close()
}

// GetServers returns the servers info from raft
func (l *DistributedLog) GetServers() ([]*api.Server, error) {
	future := l.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil, err
	}

	var servers []*api.Server
	for _, server := range future.Configuration().Servers {
		servers = append(servers, &api.Server{
			Id:       string(server.ID),
			RpcAddr:  string(server.Address),
			IsLeader: l.raft.Leader() == server.Address,
		})
	}

	return servers, nil
}

var _ raft.FSM = (*fsm)(nil)

// fsm represents the finite-state machine
type fsm struct {
	log *Log
}

// RequestType represents the request type: in this case we have only APPEND.
// We could also have DELETE, UPDATE ect..
type RequestType uint8

const (
	AppendRequestType RequestType = 0
)

// Apply represent a method to satisfy the raft RSM interface.
// Apply is called once a log entry is committed by a majority of the cluster.
func (l *fsm) Apply(record *raft.Log) interface{} {
	buf := record.Data
	reqType := RequestType(buf[0])
	switch reqType {
	case AppendRequestType:
		return l.applyAppend(buf[1:])
	}
	return nil
}

// Snapshot peridically this method to spashot its state
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	r := f.log.Reader()
	return &snapshot{reader: r}, nil
}

// Restore raft use this to restore the fsm from snapshot
func (f *fsm) Restore(r io.ReadCloser) error {
	b := make([]byte, lenWidth)
	var buf bytes.Buffer
	for i := 0; ; i++ {
		_, err := io.ReadFull(r, b)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		size := int64(enc.Uint64(b))
		if _, err = io.CopyN(&buf, r, size); err != nil {
			return err
		}
		record := &api.Record{}
		if err = proto.Unmarshal(buf.Bytes(), record); err != nil {
			return err
		}
		if i == 0 {
			f.log.Config.Segment.InitialOffset = record.Offset
			if err := f.log.Reset(); err != nil {
				return err
			}
		}
		if _, err = f.log.Append(record); err != nil {
			return err
		}
		buf.Reset()
	}
	return nil
}

// applyAppend actually applies the command stored in the finite-state machine to wal
func (l *fsm) applyAppend(b []byte) interface{} {
	var req api.ProduceRequest
	err := proto.Unmarshal(b, &req)
	if err != nil {
		return err
	}
	offset, err := l.log.Append(req.Record)
	if err != nil {
		return err
	}
	return &api.ProduceResponse{Offset: int64(offset)}
}

// FSMSnapshot inteface
var _ raft.FSMSnapshot = (*snapshot)(nil)

// snapshot implement raft.FSMSnapshot interface
type snapshot struct {
	reader io.Reader
}

// Persist, raft calls Persit on he FMSSnapshot to write its state
func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s.reader); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *snapshot) Release() {}

// raft.LogStore interface
var _ raft.LogStore = (*logStore)(nil)

// logStore is where raft store its log.
// logStore implement raft.LogStore inteface
type logStore struct {
	*Log
}

// newLogStore create a logStore
func newLogStore(dir string, config Config) (*logStore, error) {
	log, err := NewLog(dir, config)
	if err != nil {
		return nil, err
	}
	return &logStore{Log: log}, nil
}

// FirstIndex returns first index in the log
func (l *logStore) FirstIndex() (uint64, error) {
	return l.LowestOffset()
}

// LastIndex returns the last index in the log
func (l *logStore) LastIndex() (uint64, error) {
	off, err := l.HighestOffset()
	if err == io.EOF || off == 0 {
		return 0, nil
	}
	return off, nil
}

// GetLog  gets a log entry at a given index.
func (l *logStore) GetLog(index uint64, out *raft.Log) error {
	in, err := l.Read(index)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return raft.ErrLogNotFound
		}
		return err
	}
	out.Data = in.Value
	out.Index = in.Offset
	out.Type = raft.LogType(in.Type)
	out.Term = in.Term
	return nil
}

func (l *logStore) StoreLog(record *raft.Log) error {
	return l.StoreLogs([]*raft.Log{record})
}

func (l *logStore) StoreLogs(records []*raft.Log) error {
	for _, record := range records {
		if _, err := l.Append(&api.Record{
			Value: record.Data,
			Term:  record.Term,
			Type:  uint32(record.Type),
		}); err != nil {
			return err
		}
	}
	return nil
}

// DeleteRange delete old records
func (l *logStore) DeleteRange(min, max uint64) error {
	return l.Truncate(max)
}

var _ raft.StreamLayer = (*StreamLayer)(nil)

// StreamLayer represent a streamlayer in our system.
// it should implement the raft.Streamlayer interface
type StreamLayer struct {
	ln              net.Listener
	serverTLSConfig *tls.Config
	peerTLSConfig   *tls.Config
}

func NewStreamLayer(
	ln net.Listener,
	serverTlsConfig *tls.Config,
	peerTlsConfig *tls.Config,
) *StreamLayer {
	return &StreamLayer{
		ln:              ln,
		serverTLSConfig: serverTlsConfig,
		peerTLSConfig:   peerTlsConfig,
	}
}

const RaftRPC = 1

// Dial implements the Dial() function of the raft.StreamLayer interface.
// Dial makes outgoing connections to other servers in the raft cluster
func (s *StreamLayer) Dial(
	addr raft.ServerAddress,
	timeout time.Duration,
) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	var conn, err = dialer.Dial("tcp", string(addr))
	if err != nil {
		return nil, err
	}

	_, err = conn.Write([]byte{byte(RaftRPC)})
	if err != nil {
		return nil, err
	}

	if s.peerTLSConfig != nil {
		conn = tls.Client(conn, s.peerTLSConfig)
	}

	return conn, err
}

// Accept implements the Accept() method of the raft.StreamLayer interface.
// it accepts incomming connection and read the byte that identifies the connection
// and create a server-side tls connection
func (s *StreamLayer) Accept() (net.Conn, error) {
	conn, err := s.ln.Accept()
	if err != nil {
		return nil, err
	}

	b := make([]byte, 1)
	_, err = conn.Read(b)
	if err != nil {
		return nil, err
	}
	if bytes.Compare([]byte{byte(RaftRPC)}, b) != 0 {
		return nil, fmt.Errorf("not a raft rpc")
	}

	if s.serverTLSConfig != nil {
		return tls.Server(conn, s.serverTLSConfig), nil
	}

	return conn, nil
}

// Close implements the Close() function of the raft.StreamLayer interface
// Close closes the listener
func (s *StreamLayer) Close() error {
	return s.ln.Close()
}

// Addr implements the Addr() function of raft.StreamLayer interface
func (s *StreamLayer) Addr() net.Addr {
	return s.ln.Addr()
}
