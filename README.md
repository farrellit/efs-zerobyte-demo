# efs-zerobyte-demo
implement a very tight write->queue->dequeue->read loop to demonstrate lack of EFS read-after-write consistency on async mountpoints

## Dependencies 
* `docker` 
* `make`
* `go` environment
* optionally, `pyhton`

## Before you launch
* change the KeyPair for the instances.  Should be optional but isn't, sorry.
* install the golang deps from readwrite.go (it's just redis)

