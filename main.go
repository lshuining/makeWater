package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/marusama/cyclicbarrier"
	"golang.org/x/sync/semaphore"
)

// h2o 水的组成，其中我们需要俩h一个o所以我们给定他们信号量，来对他们的任务进行控制。
type H2O struct {
	// 控制的h的信号量
	seaH *semaphore.Weighted
	// 控制O的信号量
	seaO *semaphore.Weighted
	// 栅栏，这里也就是重复的使用栅栏，也就是 重复栅栏。
	cyc cyclicbarrier.CyclicBarrier
}

func NewH2O() *H2O {
	return &H2O{
		// h 两个
		seaH: semaphore.NewWeighted(2),
		// o 需要一个
		seaO: semaphore.NewWeighted(1),
		// 我们要控制的循环栅栏就是3个，因为一共需要三个嘛。
		cyc: cyclicbarrier.New(3),
	}
}

// 处理h
func (o *H2O) dealH(outH func()) {
	// 将这个信号量给拿出来1，因为h充盈来2，所以会有俩线程做这个动作
	o.seaH.Acquire(context.Background(), 1)
	// 输出 h
	outH()
	// wait的意思就是不等到三个线程，我就不走
	o.cyc.Await(context.Background())
	// 走动完毕后再把资源塞进去。
	o.seaH.Release(1)
}

// 处理 o
func (o *H2O) dealO(outO func()) {
	// 氧气将信号量中的信号取出来，
	o.seaO.Acquire(context.Background(), 1)
	// 输出o
	outO()
	// 等待三个线程跟上一个函数一样意思，也不用担心用两次不行，随便用。这个函数调用几次都 OK。
	o.cyc.Await(context.Background())
	// 释放掉。
	o.seaO.Release(1)
}
func main() {
	// channel 传递信息。
	var ch chan string
	var outO = func() {
		ch <- "O"
	}
	var outH = func() {
		ch <- "H"
	}
	// 一共有 300个channel需要。
	ch = make(chan string, 300)
	// wg是为了控制这300个线程，栅栏是为了控制生成水的这个控制器，两者的作用不同哦。
	wg := new(sync.WaitGroup)
	wg.Add(300)
	h := NewH2O()
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			h.dealO(outO)
		}()
	}
	for i := 0; i < 200; i++ {
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			h.dealH(outH)
		}()
	}
	wg.Wait()
	if len(ch) != 300 {
		fmt.Println(len(ch))
		panic("❌")
	}
	s := make([]string, 3)
	for i := 0; i < 100; i++ {
		s[0] = <-ch
		s[1] = <-ch
		s[2] = <-ch
		// 这里，对于hho进行排序了，不然也不一定是hho这个顺序
		// sort.Strings(s)
		result := s[0] + s[1] + s[2]
		fmt.Println(s)
		switch result {
		case "HHH":
			fmt.Println("错误 ❌ :", result)
		case "OOH":
			fmt.Println("错误 ❌ :", result)
		case "OHO":
			fmt.Println("错误 ❌ :", result)
		case "HOO":
			fmt.Println("错误 ❌ :", result)
		case "OOO":
			fmt.Println("错误 ❌ :", result)
		default:
			fmt.Println("正确 ✅:", result)
		}
	}
}
