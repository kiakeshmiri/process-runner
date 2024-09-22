package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"os"

	prunner "github.com/kiakeshmiri/process-runner/api/protogen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
)

func NewClient() (client prunner.ProcessServiceClient, cname string, err error) {

	tlsConfig, cname, err := LoadTLSConfig("../keys/client.pem", "../keys/client-key.pem", "../keys/root.pem")
	if err != nil {
		panic(err)
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(tlsConfig))
	if err != nil {
		panic(err)
	}

	return prunner.NewProcessServiceClient(conn), cname, nil
}

func LoadTLSConfig(certFile string, keyFile string, caFile string) (credentials.TransportCredentials, string, error) {

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load client certification: %w", err)
	}

	cn := certificate.Leaf.Subject.CommonName

	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, "", fmt.Errorf("faild to read CA certificate: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, "", fmt.Errorf("faild to append the CA certificate to CA pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
	}

	return credentials.NewTLS(tlsConfig), cn, nil
}
