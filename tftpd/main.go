package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"time"

	// tftp "projects/pkg/mod/github.com/pin/tftp/v3"

	tftp "github.com/pin/tftp/v3"
)

var (
	setblocksize bool
)

// readHandler is called when client starts file download from server
func readHandler(filename string, rf io.ReaderFrom) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}

	// Set transfer size before calling ReadFrom.
	if setblocksize {
		rf.(tftp.OutgoingTransfer).SetSize(8 * 1024)
	}
	// Set cisco "ip tftp blocksize 8192" (and allow fragmented udp packets)
	// s.SetBlockSize(65456) max if needed
	// https://github.com/pin/tftp/issues/41

	raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
	ts := time.Now()
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	file.Close()
	d := time.Now().Sub(ts)
	fmt.Printf("Sent file %s (%d bytes at %s/s) to %s\n",
		filename, n,
		prettyByteSize(float64(n)/(d.Seconds())),
		raddr.IP.String())
	return nil
}

// writeHandler is called when client starts file upload to server
func writeHandler(filename string, wt io.WriterTo) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	raddr := wt.(tftp.IncomingTransfer).RemoteAddr()
	ts := time.Now()
	n, err := wt.WriteTo(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	file.Close()
	d := time.Now().Sub(ts)
	fmt.Printf("Received file %s (%d bytes at %s/s) from %s\n",
		filename, n,
		prettyByteSize(float64(n)/(d.Seconds())),
		raddr.IP.String())
	return nil
}

func main() {
	// If you cant set larger blocksize
	setblocksize = (len(getenv("TFTP_DONTSET_BLOCKSIZE", "")) == 0)

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

func prettyByteSize(bf float64) string {
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}
