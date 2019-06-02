/*
Package github.com/olomix/chief elects a leader using etcd and runs a pool
of workers on leader. If leadership is lost, workers shut down and run on
another instance that takes leadership.
*/
package chief
