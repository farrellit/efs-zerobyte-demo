package main

import (
 "fmt"
 "os"
 "math/rand"
 "github.com/go-redis/redis"
 "strconv"
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
    base = "efs/"
  }
  go func(){
    rclient := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:6379", server), Password: "", DB: 1})
    for i := 0; passes == 0 || i < passes; i++ {
      doWrite(base, rclient, "fileq", sync);
    }
    done <- 1
  }()
  go func(){
    rclient := redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:6379", server), Password: "", DB: 1})
    for i := 0; passes == 0 || i < passes; i++ {
      doRead(rclient, "fileq");
    }
    done <- 1
  }()
  // wait for reader and writer
  _ = <-done
  _ = <-done
  fmt.Fprintln(os.Stderr, "Exiting normally")
}
func doWrite(base string, rclient *redis.Client, qkey string, sync bool) {
    fname := fmt.Sprintf("%s/%d", base, rand.Uint64())
    f, err := os.Create(fname)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Failed to create %s: %s\n", fname, err)
      return
    }
    f.WriteString(fname)
    if sync {
      if err = f.Sync(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to sync %s: %s\n", fname, err );
      }
    }
    f.Close()
    rclient.LPush(qkey, fname)
}

func doRead(rclient *redis.Client, qkey string) {
    res, err := rclient.BRPop(0, qkey).Result()
    if err != nil {
      fmt.Fprintf(os.Stderr, "Read: failed to pop from queue %s: %s\n", qkey, err)
      return
    }
    fname := res[1]
    f, err := os.Open(fname)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Read: Failed to open %s: %s\n", fname, err)
      return
    }
		b := make([]byte, 1)
    num, err := f.Read(b)
		if err != nil {
      fmt.Fprintf(os.Stderr, "Read: Failed to read from %s: %s\n", fname, err)
      f.Close()
      return
		}
		if num == 0 {
			fmt.Fprintf(os.Stderr, "Zero byte file! %s\n", fname)
      f.Close()
      return
		}
    os.Remove(fname)
    f.Close()
}
