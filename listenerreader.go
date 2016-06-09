package listenerreader

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"golang.org/x/net/context"
)

// ListenerReader wraps a net.Listener, turning it into an io.ReadCloser
type ListenerReader struct {
	bufMaxSize   int
	bufStartSize int
	cancel       context.CancelFunc
	ctx          context.Context
	delim        byte
	inputChan    chan []byte
	listener     net.Listener
}

// NewListenerReader makes a new ListenerReader.
func NewListenerReader(listener net.Listener, delim byte, bufStartSize, bufMaxSize, chanLen int) io.ReadCloser {
	ctx, cancel := context.WithCancel(context.Background())

	var inputChan chan []byte

	if chanLen < 1 {
		inputChan = make(chan []byte)
	} else {
		inputChan = make(chan []byte, chanLen)
	}

	lr := &ListenerReader{
		bufMaxSize:   bufMaxSize,
		bufStartSize: bufStartSize,
		cancel:       cancel,
		ctx:          ctx,
		delim:        delim,
		inputChan:    inputChan,
		listener:     listener,
	}

	go lr.accepter()

	return lr
}

// Read implements io.Reader. Read blocks until data is available.
func (l *ListenerReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, fmt.Errorf("Cannot read into slice of len 0")
	}

	line := <-l.inputChan

	if len(line) == 0 {
		return 0, nil
	}

	// TODO: case where len(line) > len(p)

	copied := copy(p, line)

	if copied != len(line) {
		return copied, fmt.Errorf("Error while copying, copied %d bytes out of %d line bytes", copied, len(line))
	}

	return copied, nil
}

// Close implements io.Closer
func (l *ListenerReader) Close() error {
	// TODO: shut down everything we span off into goroutines

	// close(l.inputChan) // FIXME: if any conns are still open and writing this may cause them to panic, should use a lock

	return l.listener.Close()
}

func (l *ListenerReader) accepter() {
	for {
		conn, err := l.listener.Accept()

		// TODO: error case
		if err == nil {
			go l.handler(conn)
		}
	}
}

func (l *ListenerReader) handler(conn net.Conn) {
	defer conn.Close()

	s := bufio.NewScanner(conn)
	// s.Buffer(make([]byte, l.bufStartSize), l.bufMaxSize)

	for s.Scan() {
		line := s.Bytes()
		if len(line) > 0 {
			// line is a slice, so it's a pointer to a region of memory, which
			// may be overwritten by future Scans, so we have to copy it before
			// putting it onto the chan
			// TODO: can this be done without actually copying? like passing the underlying array by value?
			foo := make([]byte, len(line)+1)
			copy(foo, line)
			foo[len(line)] = l.delim
			l.inputChan <- foo
		}
	}

	if err := s.Err(); err != nil && err != io.EOF {
		// TODO: log somehow
	}
}
