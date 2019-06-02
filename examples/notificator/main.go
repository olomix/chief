package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/olomix/chief"
)

type leaderNotifier struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

func (ln *leaderNotifier) Start() {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	if ln.cancel != nil {
		return
	}

	var ctx context.Context
	ctx, ln.cancel = context.WithCancel(context.Background())
	go ln.run(ctx.Done())
	log.Print("leader worker started")
}

func (ln *leaderNotifier) run(doneCh <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		select {
		case <-ticker.C:
		default:
		}
	}()

	for {
		select {
		case <-doneCh:
			return
		case <-ticker.C:
			log.Print("leader worker running...")
		}
	}
}

func (ln *leaderNotifier) Stop() {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.cancel == nil {
		return
	}

	ln.cancel()
	ln.cancel = nil
	log.Print("leader worker stopped")
}

func main() {
	killCh := make(chan os.Signal)
	signal.Notify(killCh, syscall.SIGTERM)
	signal.Notify(killCh, syscall.SIGINT)

	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			"http://127.0.0.1:2379",
			"http://127.0.0.1:12379",
			"http://127.0.0.1:22379",
		},
		DialTimeout:      2 * time.Second,
		AutoSyncInterval: 5 * time.Minute,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Print(err)
		}
		log.Print("etcd client closed")
	}()

	ctrl := chief.NewController(cli)
	defer func() {
		ctrl.Close()
		log.Print("leader controller closed")
	}()
	ctrl.Register(&leaderNotifier{})

	<-killCh
}
