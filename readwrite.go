package main

import (
 "fmt"
 "os"
 "math/rand"
 "github.com/go-redis/redis"
 "strconv"
 "time"
)

func main() {
  var base string
  done := make(chan int, 0)
  var sync bool
  if os.Getenv("SYNC") == "" {
    sync = false
  } else {
    sync = true
  }
  server := os.Getenv("REDIS_SERVER")
  if server == "" {
    server = "localhost"
  }
  passstr := os.Getenv("PASSES")
  passconv, err := strconv.ParseInt(passstr, 10,32)
  if err != nil {
    passconv = 0
  }
  passes := int(passconv)
  fmt.Fprintf(os.Stderr, "Starting read/write daemons, %d passes, and redis queue on %s db 1\n", passes, server)
  if len(os.Args) > 1 {
    base = os.Args[1]
  } else {
    base = "efs"
  }
  go func(){
    rclient := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:6379", server), Password: "", DB: 1})
    for i := 0; passes == 0 || i < passes; i++ {
      l, err := rclient.LLen("fileq").Result()
      if err != nil {
        fmt.Fprintf(os.Stderr, "Couldn't check length of queue to throttle writes: %s\n", err)
        continue
      } else {
        //fmt.Fprintf(os.Stderr, "Length of queue is %d\n", l)
      }
      if l > 1000 { // don't write too much, it just gobs things up
        time.Sleep( 10 * time.Second )
        i--;
        continue
      }
      doWrite(base, rclient, "fileq", sync);
      //fmt.Fprintf(os.Stderr, "doWrite done for this pass(%d/%d)\n", i,passes)
    }
    done <- 1
  }()
  go func(){
    rclient := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:6379", server), Password: "", DB: 1})
    for i := 0; passes == 0 || i < passes; i++ {
      doRead(rclient, "fileq", 0);
    }
    // drain the queue 
    for ; doRead(rclient, "fileq", 1) == true; { }
    done <- 1
  }()
  // wait for reader and writer
  _ = <-done
  _ = <-done
  fmt.Fprintln(os.Stderr, "Exiting normally")
}

func doWrite(base string, rclient *redis.Client, qkey string, sync bool) {
    fname := fmt.Sprintf("%s/%d.%d", base, rand.Uint64(), rand.Uint64() )
    f, err := os.Create(fname)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Failed to create %s: %s\n", fname, err)
      return
    }
    n, err := f.WriteString(fname)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Failed to create/write %s: %s\n", fname, err)
      return
    } else {
      if n != len(fname) {
        fmt.Fprintf(os.Stderr, "Warning, the data written to %s seems not what was expected.  Thought it would be %d but it was %d\n", fname, len(fname), n )
      }
    }
    if sync {
      if err = f.Sync(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to sync %s: %s\n", fname, err );
        f.Close()
        return
      }
    }
    f.Close()
    //fmt.Fprintf(os.Stderr, "Wrote %s and closed.\n", fname)
    _, err = rclient.LPush(qkey, fname).Result()
    if err != nil {
      fmt.Fprintf(os.Stderr, "Failed to queue %s: %s\n", fname ,err)
    }
    //fmt.Fprintf(os.Stderr, "Queued %s\n", fname)
}

func doRead(rclient *redis.Client, qkey string, timeout int64) bool {
    res, err := rclient.BRPop(time.Duration(timeout * int64(time.Second)), qkey).Result()
    if err != nil {
      fmt.Fprintf(os.Stderr, "Read: failed to pop from queue %s: %s\n", qkey, err)
      return false
    }
    fname := res[1]
    f, err := os.Open(fname)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Read: Failed to open %s: %s\n", fname, err)
      return true
    }
		b := make([]byte, 1)
    num, err := f.Read(b)
		if err != nil {
      fmt.Fprintf(os.Stderr, "Read: Failed to read from %s: %s\n", fname, err)
      f.Close()
      return true
		}
		if num == 0 {
			fmt.Fprintf(os.Stderr, "Zero byte file! %s\n", fname) // actually, we get an EOF error.
      f.Close()
      return true
		}
    f.Close()
    os.Remove(fname)
    return true
}
