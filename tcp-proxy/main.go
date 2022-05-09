package main
import (
	"fmt"
	"io"
	"net"
	"sync"
)
func proxyConn(source, destination string) error {
	connSource, err := net.Dial("tcp", source)
	if err != nil {
		return err
	}
	defer connSource.Close()
	connDestination, err := net.Dial("tcp", destination)
	if err != nil {
		return err
	}
	defer connDestination.Close()
	// connDestination replies to connSource
	 go func() { _, _ = io.Copy(connSource, connDestination) }()
	// connSource messages to connDestination
	_, err = io.Copy(connDestination, connSource)
	return err
}

func proxy1(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer)
	toReader, toIsReader := to.(io.Reader)
	if toIsReader && fromIsWriter {
		// Send replies since "from" and "to" implement the
		// necessary interfaces.
		go func() { _, _ = io.Copy(fromWriter, toReader) }()
	}
	_, err := io.Copy(to, from)
	return err
}

func main()  {
	var wg sync.WaitGroup
	server, err := net.Listen("tcp", "127.0.0.1:")
	proxyServer, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		fmt.Println(err)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := proxyServer.Accept()
			if err != nil {
				return
			}
			go func(from net.Conn) {
				defer from.Close()
				to, err := net.Dial("tcp",
					server.Addr().String())
				if err != nil {
					fmt.Println(err)
					return
				}
				defer to.Close()
				err = proxy1(from, to)
				if err != nil && err != io.EOF {
					fmt.Println(err)
				}
			}(conn)
		}
	}()

}