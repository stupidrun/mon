package middlewares

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"net"
)

type clientIPKey struct{}

func clientIP(ctx context.Context) (string, bool) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", false
	}
	addr, ok := p.Addr.(*net.TCPAddr)
	if !ok {
		return "", false
	}
	return addr.IP.String(), true
}

func IPExtractorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ip, ok := clientIP(ctx)
	if !ok {
		return nil, errors.New("could not extract client IP from context")
	}
	ctx = context.WithValue(ctx, clientIPKey{}, ip)
	return handler(ctx, req)
}

func GetClientIP(ctx context.Context) (string, bool) {
	ip, ok := ctx.Value(clientIPKey{}).(string)
	if !ok {
		return "", false
	}
	return ip, true
}
