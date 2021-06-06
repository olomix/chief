package chief

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type testWorker struct {
	m              sync.Mutex
	gotLeadership  int
	lostLeadership int
}

func (t *testWorker) Start() {
	t.m.Lock()
	defer t.m.Unlock()
	t.gotLeadership++
}

func (t *testWorker) Stop() {
	t.m.Lock()
	defer t.m.Unlock()
	t.lostLeadership++
}

func TestNewController(t *testing.T) {
	etcdAddr, ok := os.LookupEnv("ETCD_ADDR")
	require.True(t, ok)
	cfg := clientv3.Config{Endpoints: []string{etcdAddr}}
	cli, err := clientv3.New(cfg)
	defer func() {
		if err := cli.Close(); err != nil {
			t.Error(err)
		}
	}()
	members, err := cli.MemberList(context.Background())
	require.NoError(t, err)
	for _, m := range members.Members {
		t.Logf("%v %v %v", m.ID, m.Name, m.IsLearner)
	}
	require.NoError(t, err)
	ctrl := NewController(cli, "test")
	ctrl.Close()
}
