package chief

// Worker represents a worker coroutine that can be registered with
// Controller.Register. When a controller instance becomes a leader in an
// etcd cluster, it runs the worker using Worker.Start method. Start method
// should exit immediately. If it needs to run a goroutine, it should be
// started inside Start method and return as soon as possible.
//
// If leadership is lost due to etcd cluster failure or connectivity
// issues, the worker will be stopped via its Stop method.
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
