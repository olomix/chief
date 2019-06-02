package chief

// Worker represents a worker coroutine that can be registered with
// Controller.Register. When controller instance become a leader in an etcd
// cluster, it runs worker using Worker.Start method. Start method
// should exit immediately. If it needs to run gorouting, it should be run
// inside Start method and return as soon as possible.
//
// If leadership is lost as a result of etcd cluster becomes broken or
// connectivity issues, worker will be stopped with a Stop method.
type Worker interface {
	// Start starts a worker. Method is called synchronously, so if there
	// is a long job to be done, it should be run as goroutine inside Start
	// method.
	Start()

	// Stop should stop all goroutines started by Start method. And clean
	// all resources.
	Stop()
}

type Controller interface {
	// Register a Worker. If leadership is already taken, worker starts
	// immediately. Otherwise, it will be start when controller will be
	// elected as leader.
	Register(worker Worker)

	// Close stops all workers and resign leadership.
	Close()
}
