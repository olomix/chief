package chief

import (
	"context"
	"log"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type leaderCtrl struct {
	mu                sync.Mutex
	registeredWorkers []Worker
	runningWorkers    []Worker
	isLeader          bool
	wg                *sync.WaitGroup
	ctx               context.Context
	ctxCancel         context.CancelFunc
	etcdCli           *clientv3.Client
	token             string
}

// NewController returns a new Controller from etcd client.
//
// It runs worker goroutine that should be stopped using Controller.Stop
// method.
//
// token is a unique identifier of election participants group. Only one
// of participants would be elected in a group, identified by token
// on current etcd cluster.
func NewController(cli *clientv3.Client, token string) Controller {
	lc := &leaderCtrl{
		wg:      &sync.WaitGroup{},
		etcdCli: cli,
		token:   token,
	}
	lc.ctx, lc.ctxCancel = context.WithCancel(context.Background())
	lc.wg.Add(1)
	go lc.worker()
	return lc
}

func (ctrl *leaderCtrl) Register(worker Worker) {
	ctrl.mu.Lock()
	defer ctrl.mu.Unlock()
	ctrl.registeredWorkers = append(ctrl.registeredWorkers, worker)
	if ctrl.isLeader {
		worker.Start()
		ctrl.runningWorkers = append(ctrl.runningWorkers, worker)
	}
}

func (ctrl *leaderCtrl) startAll() {
	ctrl.mu.Lock()
	defer ctrl.mu.Unlock()

	if ctrl.isLeader {
		return
	}

	for _, w := range ctrl.registeredWorkers {
		w.Start()
		ctrl.runningWorkers = append(ctrl.runningWorkers, w)
	}
	ctrl.isLeader = true
}

func (ctrl *leaderCtrl) stopAll() {
	ctrl.mu.Lock()
	defer ctrl.mu.Unlock()

	if !ctrl.isLeader {
		return
	}

	for _, w := range ctrl.runningWorkers {
		w.Stop()
	}
	ctrl.runningWorkers = ctrl.runningWorkers[:0]
	ctrl.isLeader = false
}

func workerSession(ctx context.Context, cli *clientv3.Client, ctrl *leaderCtrl) error {
	// In case of closing, we need a way to interrupt NewSession creation when
	// etcd cluster is broken
	sctx, scancel := context.WithCancel(cli.Ctx())
	newSessionDoneCh := make(chan struct{})

	// goroutine to monitor session closing
	go func() {
		select {
		case <-newSessionDoneCh:
		case <-ctx.Done():
			scancel()
		}
	}()

	sess, err := concurrency.NewSession(
		cli,
		concurrency.WithTTL(10),
		concurrency.WithContext(sctx),
	)
	close(newSessionDoneCh)
	if err != nil {
		return err
	}
	defer func() {
		if err := sess.Close(); err != nil {
			log.Printf("session closed with error: %+v", err)
		}
		scancel()
	}()

	// Notify session done
	go func() {
		select {
		case <-sess.Done():
			log.Print("session done watcher")
		}
	}()

	election := concurrency.NewElection(sess, ctrl.token)
	err = election.Campaign(ctx, "")
	if err != nil {
		return err
	}

	// Ensure session is active before running registered workers
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sess.Done():
		return nil
	default:
	}

	ctrl.startAll()
	defer func() {
		ctrl.stopAll()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-sess.Done():
		return nil
	}
}

func (ctrl *leaderCtrl) worker() {
	defer ctrl.wg.Done()

	for {
		err := workerSession(ctrl.ctx, ctrl.etcdCli, ctrl)
		if err != nil {
			log.Printf("worker error: %+v", err)
		}

		// Exit on context closed
		select {
		case <-ctrl.ctx.Done():
			return
		default:
		}
	}
}

func (ctrl *leaderCtrl) Close() {
	ctrl.ctxCancel()
	ctrl.wg.Wait()
}
