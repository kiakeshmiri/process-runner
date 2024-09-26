package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	port = "50051"
)

var capool *x509.CertPool

func RunGRPCServer(registerServer func(server *grpc.Server)) {

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("server running on %s \n", port)
	RunGRPCServerOnAddr(addr, registerServer)
}

func RunGRPCServerOnAddr(addr string, registerServer func(server *grpc.Server)) {

	tlsConfig, err := LoadTLSConfig("../keys/server.pem", "../keys/server-key.pem", "../keys/root.pem")
	if err != nil {
		panic(err)
	}

	authInterceptor := NewAuthInterceptor()

	opts := []grpc.ServerOption{
		grpc.Creds(tlsConfig),
		grpc.UnaryInterceptor(authInterceptor.UnaryInterceptor),
		grpc.StreamInterceptor(authInterceptor.StreamInterceptor),
	}

	grpcServer := grpc.NewServer(opts...)

	registerServer(grpcServer)

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(grpcServer.Serve(listen))
}

func LoadTLSConfig(certFile string, keyFile string, caFile string) (credentials.TransportCredentials, error) {

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certification: %w", err)
	}

	data, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("faild to read CA certificate: %w", err)
	}

	capool = x509.NewCertPool()
	if !capool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("unable to append the CA certificate to CA pool")
	}

	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    capool,
	}
	return credentials.NewTLS(tlsConfig), nil
}
