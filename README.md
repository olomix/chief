# Chief

[![GoDoc](https://godoc.org/github.com/olomix/chief?status.svg)](https://godoc.org/github.com/olomix/chief)

Package github.com/olomix/chief elects a leader using etcd and runs a pool
of workers on leader. If leadership is lost, workers shut down and run on
another instance that takes leadership.

## TODOs

* [ ] Let `Worker.Start` return an error; if startup fails, relinquish
      leadership so another instance can try.
* [ ] Support multiple independent elections, one per worker group.
* [ ] API to query the current leader.
* [ ] API to voluntarily relinquish leadership.
* [ ] API to stop and unregister an individual worker.
