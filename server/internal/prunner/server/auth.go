package server

import (
	"context"
	"errors"
	"slices"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/kiakeshmiri/process-runner/lib/domain/clients"
)

type AuthInterceptor struct {
	accessibleRoles map[string][]string
}

type streamWrapper struct {
	grpc.ServerStream
}

func (s *streamWrapper) Context() context.Context {
	ctx := s.ServerStream.Context()

	clientID := getClientID(ctx)
	newCtx := context.WithValue(ctx, &clients.ClientContext{}, clientID)

	return newCtx
}

func NewAuthInterceptor() *AuthInterceptor {
	//Authorization table
	accessibleRoles := map[string][]string{
		"Client1": {"Start", "Stop", "GetStatus", "GetLogs"},
		"Client2": {"Stop", "GetStatus", "GetLogs"},
	}

	return &AuthInterceptor{accessibleRoles}
}

func (interceptor *AuthInterceptor) UnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	clientID, err := interceptor.authorize(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}

	newCtx := context.WithValue(ctx, &clients.ClientContext{}, clientID)

	return handler(newCtx, req)
}

func (interceptor *AuthInterceptor) StreamInterceptor(
	server interface{},
	serverStream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {

	err := handler(server, &streamWrapper{ServerStream: serverStream})

	if err != nil {
		return err
	}
	return nil
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context, method string) (string, error) {
	clientID := getClientID(ctx)

	if clientID == "" {
		return "", status.Error(codes.Unauthenticated, "unauthorized")
	}

	err := interceptor.checkAuthorization(clientID, method[strings.LastIndex(method, "/")+1:])
	if err != nil {
		return "", err
	}
	return clientID, nil
}

func (interceptor *AuthInterceptor) checkAuthorization(caller string, command string) error {
	if client, exists := interceptor.accessibleRoles[caller]; exists {
		if slices.Contains(client, command) {
			return nil
		}
	}
	return errors.New("not authorized")
}

func getClientID(ctx context.Context) string {
	var clientID string
	if p, ok := peer.FromContext(ctx); ok {
		if mtls, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			for _, item := range mtls.State.PeerCertificates {
				if item.Subject.CommonName != "" {
					clientID = item.Subject.CommonName
				}
			}
		}
	}
	return clientID
}
