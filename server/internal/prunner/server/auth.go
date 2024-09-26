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
)

type AuthInterceptor struct {
	accessibleRoles map[string][]string
}

func NewAuthInterceptor() *AuthInterceptor {
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

	err := interceptor.authorize(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func (interceptor *AuthInterceptor) StreamInterceptor(
	server interface{},
	serverStream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {

	err := interceptor.authorize(serverStream.Context(), info.FullMethod)
	if err != nil {
		return err
	}

	return handler(server, serverStream)
}

func (interceptor *AuthInterceptor) authorize(ctx context.Context, method string) error {
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

	if clientID == "" {
		return status.Error(codes.Unauthenticated, "unauthorized")
	}

	return interceptor.checkAuthorization(clientID, method[strings.LastIndex(method, "/")+1:])
}

func (interceptor *AuthInterceptor) checkAuthorization(caller string, command string) error {
	if client, exists := interceptor.accessibleRoles[caller]; exists {
		if slices.Contains(client, command) {
			return nil
		}
	}
	return errors.New("not authorized")
}
