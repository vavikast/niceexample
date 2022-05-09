package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

type Pxy struct {}

func (p *Pxy)ServeHTTP(w http.ResponseWriter,r *http.Request)  {
	fmt.Printf("Received request %s %s %s\n", r.Method, r.Host, r.RemoteAddr)
	transport := http.DefaultTransport

	//step1
	outReq := new(http.Request)
	*outReq = *r
	if clientIP,_ ,err := net.SplitHostPort(r.RemoteAddr); err == nil{
		if prior,ok := outReq.Header["X-Forwarded-For"]; ok{
			clientIP = strings.Join(prior,",")+", "+clientIP
		}
		outReq.Header.Set("X-Forwarded-For",clientIP)
	}
	res, err := transport.RoundTrip(outReq)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	//step3
	for k,v := range r.Header{
		for _, v := range v {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(res.StatusCode)
	io.Copy(w,res.Body)
	res.Body.Close()



}

func main() {
	fmt.Println("Serve on :8080")
	http.Handle("/", &Pxy{})
	http.ListenAndServe("0.0.0.0:8080", nil)
}