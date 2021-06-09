package chief

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	require.NoError(t, err)
	defer func() {
		if err := cli.Close(); err != nil {
			t.Error(err)
		}
	}()
	ctrl := NewController(cli, "test")
	worker := &testWorker{}
	ctrl.Register(worker)

	start := time.Now()
	for {
		worker.m.Lock()
		x := worker.gotLeadership
		worker.m.Unlock()
		if x > 0 {
			break
		}
		require.True(t, time.Now().Before(start.Add(30*time.Second)))
		time.Sleep(200 * time.Millisecond)
	}

	ctrl.Close()

	worker.m.Lock()
	x := worker.gotLeadership
	worker.m.Unlock()
	assert.Equal(t, 1, x)
}
