package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		_, _ = fmt.Fprintln(os.Stderr, "usage: go-telnet [--timeout=10s] host port")
		os.Exit(2)
	}
	host := args[0]
	port := args[1]
	addr := net.JoinHostPort(host, port)

	client := NewTelnetClient(addr, timeout, os.Stdin, os.Stdout)
	if err := client.Connect(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "...Failed to connect to %s: %v\n", addr, err)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stderr, "...Connected to %s\n", addr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	sendErrCh := make(chan error, 1)
	recvErrCh := make(chan error, 1)

	go func() { sendErrCh <- client.Send() }()
	go func() { recvErrCh <- client.Receive() }()

	for {
		select {
		case <-ctx.Done():
			stop()
			_ = client.Close()
			_, _ = fmt.Fprintln(os.Stderr, "...Interrupted")
			return

		case err := <-sendErrCh:
			_ = client.Close()
			if err == nil {
				_, _ = fmt.Fprintln(os.Stderr, "...EOF")
				return
			}
			_, _ = fmt.Fprintf(os.Stderr, "...Send error: %v\n", err)
			os.Exit(1)

		case err := <-recvErrCh:
			_ = client.Close()
			if err == nil || errors.Is(err, net.ErrClosed) {
				_, _ = fmt.Fprintln(os.Stderr, "...Connection was closed by peer")
				return
			}
			stop()
			_, _ = fmt.Fprintf(os.Stderr, "...Receive error: %v\n", err)
			os.Exit(1)
		}
	}
}
