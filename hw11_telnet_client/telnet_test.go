package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("connect error", func(t *testing.T) {
		// Listen on a random port and immediately close it, then ensure connect fails.
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		addr := l.Addr().String()
		require.NoError(t, l.Close())

		client := NewTelnetClient(addr, 50*time.Millisecond, io.NopCloser(bytes.NewBuffer(nil)), io.Discard)
		err = client.Connect()
		require.Error(t, err)
	})

	t.Run("send fails after peer hard close", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		accepted := make(chan struct{})
		go func() {
			defer close(accepted)
			conn, err := l.Accept()
			require.NoError(t, err)
			tcp := conn.(*net.TCPConn)
			// Linger=0 makes close send RST, so client write should error deterministically.
			require.NoError(t, tcp.SetLinger(0))
			require.NoError(t, tcp.Close())
		}()

		in := bytes.NewBufferString("hello\n")
		client := NewTelnetClient(l.Addr().String(), time.Second, io.NopCloser(in), io.Discard)
		require.NoError(t, client.Connect())
		defer func() { require.NoError(t, client.Close()) }()

		<-accepted
		time.Sleep(10 * time.Millisecond)
		err = client.Send()
		require.Error(t, err)
	})
}
