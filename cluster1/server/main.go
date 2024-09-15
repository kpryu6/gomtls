package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

// grpc 서비스 구현
// helloworld.proto에 정의된 Greeter 서비스 기본 구현
type greeter struct {
	helloworld.UnimplementedGreeterServer
}

func (greeter) SayHello(ctx context.Context, req *helloworld.HelloRequest) (
	*helloworld.HelloReply, error,
) {
	clientId := "UnknownClient"

	// client's SVID
	if peerId, ok := grpccredentials.PeerIDFromContext(ctx); ok {
		clientId = peerId.String()
	}

	log.Printf("Server gets %s's SPIFFEID : %q", req.Name, clientId)

	return &helloworld.HelloReply{
		Message: fmt.Sprintf("This is Server Reply : Hello %s.", req.Name),
	}, nil
}

func main() {

	var addr string
	flag.StringVar(&addr, "addr", "localhost:8123", "host:server port")
	flag.Parse()

	log.Printf("Server (%s) starting up...", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	// establish secure communication by using SVID and Bundles from the SPIRE Workload API
	spiffeEndpointSocket := os.Getenv("SPIFFE_ENDPOINT_SOCKET")
	log.Printf("SPIFFE_ENDPOINT_SOCKET: %s", spiffeEndpointSocket)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Get source from workloadAPI
	// Source contains workload's SVID and trust bundles
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		log.Fatalf("Unable to create X509Source: %v", err)
	}
	defer source.Close()
	// 보안되지 않은 gRPC
	// creds := grpc.Creds(insecure.NewCredentials())

	// SVID cryptographically identifies the server workload.
	svid, _ := source.GetX509SVID()
	if err != nil {
		log.Fatalf("Failed to get X509-SVID: %v", err)
	}
	log.Println("Server's spiffeID: ", svid.ID)

	// 이 작업 대신 istio로 대체 가능 (how??)
	// source : workload's SVID
	// source : trust bundle
	// tlsconfig.AuthorizeAny() : 인증서가 유효하거나 신뢰하면 어떤 연결이든 허용
	// 							  엄격하게 접근제어를 하기위해선 대신 OPA로 설정 가능
	creds := grpc.Creds(
		grpccredentials.MTLSServerCredentials(
			source, source, tlsconfig.AuthorizeAny(),
		),
	)

	server := grpc.NewServer(creds)

	// greeter 서비스를 gRPC 서버에 등록
	helloworld.RegisterGreeterServer(server, greeter{})

	log.Println("Serving on", listener.Addr())
	// 서비스 제공
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}

}
