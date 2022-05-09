package main

import (
	"fmt"
	"log"
	"net/http"
)

func handle1(w http.ResponseWriter, r *http.Request) {
	hj, _ := w.(http.Hijacker)
	conn, buf, _ := hj.Hijack()
	defer conn.Close()
	buf.WriteString("HTTP/1.0 200 OK\r\n")
	buf.WriteString("Content-Type: multipart/x-mixed-replace; boundary=\"socketio\"\r\n")
	buf.WriteString("Connection: keep-alive\r\n")
	buf.WriteString("\r\n--socketio\r\n")
	buf.WriteString("hello world")
	buf.Flush()
}

func handle2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func main()  {
	http.Handle("/x",http.HandlerFunc(handle1))
	http.Handle("/y",http.HandlerFunc(handle2))
	log.Fatalln(http.ListenAndServe(":8888",nil))

}