package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/peer"
)

func sendRequest(ctx context.Context, client helloworld.GreeterClient) {
	peer := new(peer.Peer)
	res, err := client.SayHello(ctx, &helloworld.HelloRequest{
		Name: "Cluster1 Client",
	}, grpc.Peer(peer))
	if err != nil {
		log.Printf("Failed to say hello: %v\n", err)
		return
	}

	// TODO: Learn the server's identity
	serverId := "Unknownserver"
	if peerId, ok := grpccredentials.PeerIDFromPeer(peer); ok {
		serverId = peerId.String()
	}
	log.Printf("%s\n", res.Message)
	log.Printf("Client gets Server's SPIFFEID : %q", serverId)
}
func main() {
	var addr string
	flag.StringVar(&addr, "addr", "", "host:port of the server")
	flag.Parse()

	if addr == "" {
		addr = os.Getenv("GREETER_SERVER_ADDR")
	}
	if addr == "" {
		addr = "localhost:8123"
	}

	log.Println("Client starting up...", addr)

	ctx := context.Background()

	serverId := spiffeid.RequireFromString(
		"spiffe://cluster1.org/ns/default/sa/default/app/helloworld-server",
	)

	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		log.Fatalf("Unable to create X509Source: %v", err)
	}

	// serverId를 가지고 있는 아이만 허락
	creds := grpc.WithTransportCredentials(
		grpccredentials.MTLSClientCredentials(
			source, source, tlsconfig.AuthorizeID(serverId),
		))

	client, err := grpc.DialContext(ctx, addr, creds)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	greeterClient := helloworld.NewGreeterClient(client)

	const interval = time.Second * 20
	log.Printf("Will send request every %s...\n", interval)
	for {
		sendRequest(ctx, greeterClient)
		time.Sleep(interval)
	}

}
