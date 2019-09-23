package connectstream

import (
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

var eventChan = make(chan string, 100)

func eventDebug() {
	for msg := range eventChan {
		fmt.Println("NEW EVENT", msg)
	}
}

func e(a string) {
	eventChan <- a
}

func PTestWritingAndFastClosing(t *testing.T) {

	go eventDebug()

	const sendBytes = 98880

	var (
		serverReceivedTheRequest bool
		listenOn                 = make(chan string)
		errChan                  = make(chan error)
	)

	// start server
	go func() {

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			errChan <- err
		}
		listenOn <- listener.Addr().String()

		conn, err := listener.Accept()
		if err != nil {
			errChan <- err
		}

		e("started reading")

		bytes, err := ioutil.ReadAll(conn)
		if err != nil {
			errChan <- err
		}

		e("finished reading")

		serverReceivedTheRequest = len(bytes) == sendBytes
		errChan <- nil
	}()

	conn, err := net.Dial("tcp", <-listenOn)
	if err != nil {
		t.Fatal(err)
	}

	e("started writing")

	writeBuf := make([]byte, sendBytes)
	_, err = conn.Write(writeBuf)
	if err != nil {
		t.Fatal(err)
	}

	e("finished writing")

	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}

	e("closing")

	err = <-errChan
	if err != nil {
		t.Fatal(err)
	}

	if !serverReceivedTheRequest {
		t.Fail()
	}

}

func PTestWritingAndFastClosing2(t *testing.T) {

	go eventDebug()

	const sendBytes = 98880

	var (
		serverReceivedTheRequest bool
		errChan                  = make(chan error)

		left, right = net.Pipe()
	)

	// start server
	go func() {

		e("started reading")

		bytes, err := ioutil.ReadAll(left)
		if err != nil {
			errChan <- err
		}

		e("finished reading")

		serverReceivedTheRequest = len(bytes) == sendBytes
		errChan <- nil
	}()

	e("started writing")

	writeBuf := make([]byte, sendBytes)
	_, err := right.Write(writeBuf)
	if err != nil {
		t.Fatal(err)
	}

	e("finished writing")

	err = right.Close()
	if err != nil {
		t.Fatal(err)
	}

	e("closing")

	err = <-errChan
	if err != nil {
		t.Fatal(err)
	}

	if !serverReceivedTheRequest {
		t.Fail()
	}

}

func TestWritingAndFastClosing3(t *testing.T) {

	go eventDebug()

	const sendBytes = 98880

	var (
		serverReceivedTheRequest bool
		listenOn                 = make(chan string)
		proxyOn                  = make(chan string)
	)

	// start server
	var errServerChan = make(chan error)
	go func() {

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			errServerChan <- err
		}
		listenOn <- listener.Addr().String()

		conn, err := listener.Accept()
		if err != nil {
			errServerChan <- err
		}

		e("started reading")

		time.Sleep(time.Second)

		bytes, err := ioutil.ReadAll(conn)
		if err != nil {
			errServerChan <- err
		}

		e("finished reading")

		serverReceivedTheRequest = len(bytes) == sendBytes
		errServerChan <- nil
	}()

	// start proxy
	var errProxyChan = make(chan error)
	go func() {

		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			errProxyChan <- err
		}
		proxyOn <- listener.Addr().String()

		from, err := listener.Accept()
		if err != nil {
			errProxyChan <- err
		}

		e("proxy connecting")
		to, err := net.Dial("tcp", <-listenOn)
		if err != nil {
			errProxyChan <- err
		}

		e("started proxying")

		errProxyChan <- Connect(from, to)
	}()

	conn, err := net.Dial("tcp", <-proxyOn)
	if err != nil {
		t.Fatal(err)
	}

	e("started writing")

	writeBuf := make([]byte, sendBytes)
	_, err = conn.Write(writeBuf)
	if err != nil {
		t.Fatal(err)
	}

	e("finished writing")

	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}

	e("closing")

	err = <-errServerChan
	if err != nil {
		t.Fatal(err)
	}

	err = <-errProxyChan
	if err != nil {
		t.Fatal(err)
	}

	if !serverReceivedTheRequest {
		t.Fail()
	}

}
