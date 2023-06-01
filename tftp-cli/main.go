package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"time"

	"github.com/pin/tftp/v3"
)

// tftp "github.com/pin/tftp/v3"
var (
	myLogger = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func main() {

	var (
		put, get                      bool
		localFile, Server, remoteFile string
		blocksize                     int
	)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "[options] [hosts...]")
		flag.PrintDefaults()
	}

	flag.BoolVar(&put, "put", false, "Send")
	flag.BoolVar(&get, "get", false, "Get")

	flag.StringVar(&localFile, "l", "", "local file")
	flag.StringVar(&remoteFile, "r", "", "remote filename")
	flag.StringVar(&Server, "s", "", "tftp server address")
	flag.IntVar(&blocksize, "blocksize", 8*1024, "blocksize")

	flag.Parse()

	c, err := tftp.NewClient(Server)
	if err != nil {
		myLogger.Fatal(err)
	}
	c.SetBlockSize(blocksize)

	if len(Server) == 0 || len(remoteFile) == 0 || (put && get) {
		flag.Usage()
		os.Exit(-1)
	}

	switch {

	case put: // UPLOAD :
		file, err := os.Open(localFile)
		if err != nil {
			myLogger.Fatal(err)
		}
		ts := time.Now()

		// c.SetTimeout(3 * time.Second) // optional
		rf, err := c.Send(remoteFile, "octet")
		if err != nil {
			myLogger.Fatal(err)
		}
		n, err := rf.ReadFrom(file)
		if err != nil {
			myLogger.Fatal(err)
		}
		file.Close()
		d := time.Now().Sub(ts)

		fmt.Printf("PUT %s as %s/%s : %d bytes sent at %s/s\n",
			localFile, Server, remoteFile,
			n, prettyByteSize(float64(n)/(d.Seconds())))

	case get: // DOWNLOAD:
		// c.SetTimeout(3 * time.Second) // optional
		// c.RequestTSize(true)
		ts := time.Now()
		wt, err := c.Receive(remoteFile, "octet")
		if err != nil {
			log.Fatal(err)
		}
		if len(localFile) == 0 {
			localFile = path.Base(remoteFile)
		}
		file, err := os.Create(localFile)
		if err != nil {
			log.Fatal(err)
		}
		// Optionally obtain transfer size before actual data.
		// if n, ok := wt.(tftp.IncomingTransfer).Size(); ok {
		// 	fmt.Printf("Transfer size: %d\n", n)
		// }
		n, err := wt.WriteTo(file)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
		d := time.Now().Sub(ts)
		fmt.Printf("GET %s/%s as %s : %d bytes received at %s/s\n",
			Server, remoteFile, localFile,
			n, prettyByteSize(float64(n)/(d.Seconds())))
		file.Close()

	default:
		flag.Usage()
	}

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
