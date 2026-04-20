// Package rum implements the continous flow of concurrent funcs manager which can be regisreted via server can be called via clinet
// flow:
//
//												                Create Profile:
//											                            ↓
//											                    "profile" includes:
//											         kit: provides the descirption of the profile that includes:
//											             "services": contains descirption of the service which include:
//										                             "time-format" ->  deactivate the service, remove the service,activation time, retry call if the dispatch failed, on what time to invoke, duration of delay per dispatch,
//									                                 "dispatcher" -> controls the number of registered funcs
//								                                     "budget"     -> controls the mode budget
//													     "time-format": same as services but for the profile
//							                                  Call The Profile:
//							                                            ↓
//					                                        grpc accepts the call & triggers hub through channel
//				                                                     ↓
//			                                               hub performs as per the call:
//						                                 onPost-> fetches the service -> reads the format -> performs the write -> write publishes the work -> tickFetch fetches the result -> Paper publishes the result -> result is passed to the client
//	                                                  onDeactivate-> read desc ->  temporarly remove the service or profile
//	                                                  onActivate-> read desc ->  find the deactivate serivce or profile -> activates the service or profile
//	                                                  onRemove-> read desc ->  remove the service or profile
package rum

import (
	"context"
	"log"
	"net"
	rumrpc "rum/app/misc/rum"
	rumpaint "rum/app/paint"
	"runtime/debug"

	"google.golang.org/grpc"
)

type RumServer struct {
	Network       string
	Address       string
	ServerOptions []grpc.ServerOption
}

// Serve starts the service
func (r *Rum[In, Out]) Serve(ctx context.Context, server RumServer) {
	rumpaint.Header(`
██████╗░██╗░░░██╗███╗░░░███╗
██╔══██╗██║░░░██║████╗░████║
██████╔╝██║░░░██║██╔████╔██║
██╔══██╗██║░░░██║██║╚██╔╝██║
██║░░██║╚██████╔╝██║░╚═╝░██║
╚═╝░░╚═╝░╚═════╝░╚═╝░░░░░╚═╝
	`)
	network := server.Network
	address := server.Address
	opts := server.ServerOptions
	lis, err := net.Listen(network, address)

	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(opts...)
	rumrpc.RegisterOnRumServiceServer(grpcServer, r)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC RECOVERED in Hub goroutine: %v\n%s", r, debug.Stack())
			}
		}()
		r.Hub()
	}()

	go func() {
		<-ctx.Done()
		log.Println("shutting down gRPC server...")
		grpcServer.GracefulStop()
	}()

	log.Println("start :)")
	if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
		log.Fatalf("failed to serve: %v", err)
	}
}
