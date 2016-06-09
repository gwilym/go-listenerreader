package listenerreader

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"

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

	fmt.Fprint(os.Stderr, "LR Reading chan with p len ", len(p), "\n")
	line := <-l.inputChan

	fmt.Fprint(os.Stderr, "LR Read chan ", len(line), " bytes: ", string(line[0:64]), "\n")

	if len(line) == 0 {
		return 0, nil
	}

	// TODO: case where len(line) > len(p)

	copied := copy(p, line)

	// fmt.Fprint(os.Stderr, "LR Copied line to: ", p, "\n")

	if copied != len(line) {
		return copied, fmt.Errorf("Error while copying, copied %d bytes out of %d line bytes", copied, len(line))
	}

	return copied, nil
}

// Close implements io.Closer
func (l *ListenerReader) Close() error {
	// TODO: shut down everything we span off into goroutines

	fmt.Fprint(os.Stderr, "LR Closing\n")

	// close(l.inputChan) // FIXME: if any conns are still open and writing this may cause them to panic, should use a lock

	return l.listener.Close()
}

func (l *ListenerReader) accepter() {
	fmt.Fprint(os.Stderr, "LR Accepting ...\n")
	for {
		conn, err := l.listener.Accept()

		// TODO: error case
		if err == nil {
			fmt.Fprint(os.Stderr, "LR Handling ...\n")
			go l.handler(conn)
		}
	}
}

func (l *ListenerReader) handler(conn net.Conn) {
	defer conn.Close()

	s := bufio.NewScanner(conn)
	// s.Buffer(make([]byte, l.bufStartSize), l.bufMaxSize)

	fmt.Fprint(os.Stderr, "LR Scanning ...\n")
	for s.Scan() {
		line := s.Bytes()
		fmt.Fprint(os.Stderr, "LR Scanned ", len(line), " bytes: ", string(line[0:64]), "\n")
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
	fmt.Fprint(os.Stderr, "LR Scan done, err: ", s.Err(), "\n")

	if err := s.Err(); err != nil && err != io.EOF {
		// TODO: log somehow
	}
}
