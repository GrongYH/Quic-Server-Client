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
	hostName := flag.String("hostname", "124.160.115.141", "hostname/ip of the server")
	portNum := flag.String("port", "18466", "port number of the server")

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

	resp := make(chan string)

	var size int = 10000
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
