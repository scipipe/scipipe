package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	runtime.GOMAXPROCS(4)

	c1 := make(chan int, 16)
	c2 := make(chan int, 16)

	mx := new(sync.Mutex)

	go func() {
		for i := 1; i <= 10; i++ {
			mx.Lock()
			c1 <- i
			c2 <- i
			mx.Unlock()
		}
	}()

	go func() {
		for i := 101; i <= 110; i++ {
			mx.Lock()
			c1 <- i
			c2 <- i
			mx.Unlock()
		}
	}()

	for i := 1; i <= 20; i++ {
		i1 := <-c1
		i2 := <-c2
		fmt.Println(i1, i2)
	}
}
