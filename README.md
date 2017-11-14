# efs-zerobyte-demo
implement a very tight write->queue->dequeue->read loop to demonstrate lack of EFS read-after-write consistency on async mountpoints

## Dependencies 
* `docker` 
* `make`
* `go` environment
* optionally, `pyhton`

## Before you launch
* change the KeyPair for the instances.  Should be optional or configurable but isn't, sorry.
* install the golang deps from readwrite.go (it's just redis)

## Launching

`make stack` should do it.  

## Interacting

check the output in the CloudWatch Logs group, which will be created with the stack.  It's typically pretty quiet except for errors.

EOF Error indicates the program got EOF when it read the file - since the file ins't queued until the Close() happens, this would indicate lack of read after write consistency on unsynced files on `async` (the default) mount points.

Other errors are not anticipated.
