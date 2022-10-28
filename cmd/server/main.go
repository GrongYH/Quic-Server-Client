package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"strings"

	quic "github.com/lucas-clemente/quic-go"
)

type Request struct {
	Size int `json:"size"`
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo"},
	}
}

func main() {
	err := startQUICServer()
	if err != nil {
		panic(err)
	}
	select {}
}

func startQUICServer() (err error) {
	hostName := flag.String("hostname", "124.160.115.141", "hostname/ip of the server")
	portNum := flag.String("port", "18466", "port number of the server")

	flag.Parse()

	addr := *hostName + ":" + *portNum

	log.Println("Server running @", addr)

	listener, err := quic.ListenAddr(addr, generateTLSConfig(), nil)
	if err != nil {
		return
	}
	defer listener.Close()
	for {
		sess, err := listener.Accept(context.Background())
		if err != nil {
			break
		}
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			break
		}
		go func(stream quic.Stream) {
			defer stream.Close()
			for {
				var sb strings.Builder
				fmt.Println("start reading...")
				b := make([]byte, 1024*4)
				n, _ := stream.Read(b)
				sb.Write(b[:n])
				req := new(Request)
				println(sb.String())
				json.Unmarshal([]byte(sb.String()), req)
				fmt.Println(req)
				remain := req.Size
				if remain <= 0 {
					return
				}
				buf := make([]byte, remain)
				for i := 0; i < remain; i++ {
					buf[i] = '1'
				}
				fmt.Printf("已经发送%s", buf)
				stream.Write(buf)
			}
		}(stream)
	}
	return
}
