package main

import (
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func server(dst string) {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})
	r.Run(dst)
}
func main() {
	dst := "localhost:9091"
	go server(dst)
	time.Sleep(time.Second * 5)
	ProxyTCP(dst)
	for {

	}
}
func ProxyTCP(dst string) error {
	var wg sync.WaitGroup
	proxyServer, err := net.Listen("tcp", "127.0.0.1:9090")
	if err != nil {
		log.Fatal("err")
	}
	defer proxyServer.Close()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := proxyServer.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
			go func(from net.Conn) {
				defer from.Close()
				to, err := net.Dial("tcp", dst)
				if err != nil {
					log.Fatal(err)
				}
				defer to.Close()
				err = proxy(from, to)
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
			}(conn)
		}
	}()
	wg.Wait()
	return nil
}
func proxy(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer)

	toReader, toIsReader := to.(io.Reader)

	if toIsReader && fromIsWriter {
		go func() {
			io.Copy(fromWriter, toReader)
		}()
	}
	_, err := io.Copy(to, from)
	return err
}
