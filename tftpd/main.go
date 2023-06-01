package main

import (
	"fmt"
	"io"
	"os"
	"time"

	// tftp "projects/pkg/mod/github.com/pin/tftp/v3"

	tftp "github.com/pin/tftp/v3"
)

// readHandler is called when client starts file download from server
func readHandler(filename string, rf io.ReaderFrom) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}

	// Set transfer size before calling ReadFrom.
	if len(getenv("TFTP_LOW_BLOCKSIZE", "")) == 0 {
		rf.(tftp.OutgoingTransfer).SetSize(8 * 1024)
	}
	// Set cisco "ip tftp blocksize 8192" (and allow fragmented udp packets)
	// s.SetBlockSize(65456) max if needed
	// https://github.com/pin/tftp/issues/41

	raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("Sent file %s (%d bytes) to %s\n", filename, n, raddr.IP.String())
	return nil
}

// writeHandler is called when client starts file upload to server
func writeHandler(filename string, wt io.WriterTo) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := wt.WriteTo(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	raddr := wt.(tftp.IncomingTransfer).RemoteAddr()
	fmt.Printf("Received file %s (%d bytes) from %s\n", filename, n, raddr.IP.String())
	return nil
}

func main() {
	// use nil in place of handler to disable read or write operations
	s := tftp.NewServer(readHandler, writeHandler)
	s.SetBackoff(func(attempts int) time.Duration { // retransmit unacknowledged
		return time.Duration(attempts) * time.Second
	})
	// s.SetBackoff(func(int) time.Duration { return 0 }) // No need for retries ?
	// s.SetTimeout(5 * time.Second) // optional
	err := s.ListenAndServe(getenv("DOCKER_TFTP_PORT", ":69"))
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
