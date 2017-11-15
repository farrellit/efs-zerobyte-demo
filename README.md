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
* ecr repo should be created by default but isn't

## Launching

`make stack` should do it.  

## Interacting

check the output in the CloudWatch Logs group, which will be created with the stack.  It's typically pretty quiet except for errors.

At significant numbers of processor nodes, `No Such File or Directory` errors are common - I _believe_ they are the effect of directory caching, but I haven't had a chance to asses that yet.  

`EOF` Error indicates the program got EOF when it read the file - since the file ins't queued until _after_ the Close() happens, this would indicate lack of read after write consistency on unsynced files on `async` (the default) mount points.  This is not unexpected, based on the contents of the `nfs` manual page details regarding `sync` vs `async` mounts, but the coconverstaion thus far has been somewhat unclear on that point.

Other errors are not anticipated and have not yet been seen.
