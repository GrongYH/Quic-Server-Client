package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"io"
	"log"

	quic "github.com/lucas-clemente/quic-go"
)

type Request struct {
	Size int `json:"size"`
}

func main() {
	hostName := flag.String("hostname", "localhost", "hostname/ip of the server")
	portNum := flag.String("port", "4240", "port number of the server")

	flag.Parse()

	addr := *hostName + ":" + *portNum

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo"},
	}

	session, err := quic.DialAddr(addr, tlsConf, nil)
	if err != nil {
		panic(err)
	}

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		panic(err)
	}
	//timeout := time.Duration(*timeoutDuration) * time.Millisecond

	resp := make(chan string)

	var size int
	for size = 1; size < 1000; size++ {
		message := Request{
			Size: size,
		}
		b, err := json.Marshal(message)
		if err != nil {
			panic(nil)
		}

		log.Printf("Client: Sending '%s'\n", b)
		_, err = stream.Write(b)
		if err != nil {
			panic(err)
		}

		log.Println("Done. Waiting for echo")

		go func() {
			buff := make([]byte, size)
			_, _ = io.ReadFull(stream, buff)
			resp <- string(buff)
		}()
		select {
		case reply := <-resp:
			log.Printf("Client: Got '%s'\n", reply)
		}
	}

}
