package connectstream

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type closeWriter interface {
	CloseWrite() error
}

type closeReader interface {
	CloseRead() error
}

func responsibleCopy(a, b io.ReadWriteCloser, wg *sync.WaitGroup, errors chan<- error) {

	defer wg.Done()
	defer fmt.Println("wat")

	_, err := io.Copy(a, b)
	errors <- err

	if cw, ok := b.(closeWriter); ok {
		errors <- cw.CloseWrite()
	} else {
		errors <- b.Close()
	}

	if cw, ok := a.(closeReader); ok {
		errors <- cw.CloseRead()
	} else {
		errors <- a.Close()
	}
}

// Connect connects two streams using io.Copy
// whenever any operation returns, both streams are closed
// function returns only both Copy are finished
func Connect(a, b io.ReadWriteCloser) error {

	var (
		wg     sync.WaitGroup
		errors = make(chan error, 32)
	)

	wg.Add(2)
	go responsibleCopy(a, b, &wg, errors)
	go responsibleCopy(b, a, &wg, errors)
	wg.Wait()

	errors <- a.Close()
	errors <- b.Close()

	close(errors)

	var resultError *multierror.Error
	var closedConn = "use of closed network connection"
	var endpointNotConnected = "transport endpoint is not connected"

	for err := range errors {

		if err == nil || err == io.ErrClosedPipe {
			continue
		}

		if strings.Contains(err.Error(), closedConn) {
			continue
		}

		if strings.Contains(err.Error(), endpointNotConnected) {
			continue
		}

		resultError = multierror.Append(resultError, err)
	}

	return resultError.ErrorOrNil()
}
