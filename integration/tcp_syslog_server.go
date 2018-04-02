package integration

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/onsi/gomega/gbytes"
)

type TcpSyslogServer struct {
	Addr   string
	Buffer *gbytes.Buffer
}

func (s *TcpSyslogServer) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	// Listen for incoming connections.
	l, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + s.Addr)

	close(ready)

	var conn net.Conn

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err = l.Accept()
			if err != nil {
				return
			}

			_, err = io.Copy(s.Buffer, conn)

			// io.Copy is blocking. So when we close the underlying connection after
			// being signalled, we need to check for that error
			if err != nil {
				newErr, ok := err.(*net.OpError)
				if ok {
					if strings.Contains(newErr.Error(), "use of closed network connection") {
						return
					}
				}
				panic(err)
			}

			conn.Close()
		}
	}()

	<-signals
	if conn != nil {
		conn.Close()
	}

	return nil
}
