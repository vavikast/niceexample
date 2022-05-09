
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 监听全部信号
func main()  {
	//合建chan
	c := make(chan os.Signal)
	//监听所有信号
	signal.Notify(c, syscall.SIGINT, syscall.SIGILL, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGSEGV, syscall.SIGTERM)
	//阻塞直到有信号传入
	fmt.Println("启动")
	s := <-c
	fmt.Println("退出信号: ", s)
	fmt.Println("x")
	time.Sleep(time.Second)

}