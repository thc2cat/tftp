package main

// V0.4 : Adding timestamp in logs.
//        Better Dockerfile With multibuild and TZ set
// V0.41 : Thinner container with TZ Loading

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"time"

	tftp "github.com/pin/tftp/v3"
)

var (
	setBlockSize bool
)

const (
	YYYYMMDD  = "2006-01-02"
	HHMMSS24h = "15:04:05"
)

// readHandler is called when the client starts file download from the server
func readHandler(filename string, rf io.ReaderFrom) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set cisco "ip tftp blocksize 8192" (and allow fragmented udp packets)
	// s.SetBlockSize(65456) max if needed
	// https://github.com/pin/tftp/issues/41

	// Set transfer size before calling ReadFrom.
	if setBlockSize {
		rf.(tftp.OutgoingTransfer).SetSize(8 * 1024)
	}

	raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
	startTime := time.Now()
	n, err := rf.ReadFrom(file)
	if err != nil {
		return err
	}

	elapsedTime := time.Since(startTime)
	datetime := time.Now().Local().Format(YYYYMMDD + " " + HHMMSS24h)
	fmt.Printf("%s: Sent %s (%d bytes at %s/s) to %s\n",
		datetime,
		filename, n,
		prettyByteSize(float64(n)/(elapsedTime.Seconds())),
		raddr.IP.String())
	return nil
}

// writeHandler is called when the client starts file upload to the server
func writeHandler(filename string, wt io.WriterTo) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	raddr := wt.(tftp.IncomingTransfer).RemoteAddr()
	startTime := time.Now()
	n, err := wt.WriteTo(file)
	if err != nil {
		return err
	}

	elapsedTime := time.Since(startTime)
	datetime := time.Now().Local().Format(YYYYMMDD + " " + HHMMSS24h)
	fmt.Printf("%s: Received %s (%d bytes at %s/s) from %s\n",
		datetime,
		filename, n,
		prettyByteSize(float64(n)/(elapsedTime.Seconds())),
		raddr.IP.String())
	return nil
}

func main() {
	// If you can't set a larger blocksize
	setBlockSize = len(getenv("TFTP_DONTSET_BLOCKSIZE", "")) == 0

	// Use nil in place of a handler to disable read or write operations
	s := tftp.NewServer(readHandler, writeHandler)
	s.SetBackoff(func(attempts int) time.Duration {
		return time.Duration(attempts) * time.Second // Retransmit unacknowledged
	})
	// s.SetBackoff(func(int) time.Duration { return 0 }) // No need for retries ?
	// s.SetTimeout(5 * time.Second) // optional

	// manually set time zone ( Using scratch Docker Container )
	if tz := os.Getenv("TZ"); tz != "" {
		var err error
		time.Local, err = time.LoadLocation(tz)
		if err != nil {
			log.Printf("error loading location '%s': %v\n", tz, err)
		}
	}

	err := s.ListenAndServe(getenv("DOCKER_TFTP_PORT", ":69"))
	if err != nil {
		fmt.Printf("server: %v\n", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if exists {
		return value
	}
	return fallback
}

func prettyByteSize(bf float64) string {
	units := []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"}
	for _, unit := range units {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}
