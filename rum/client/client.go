// Package client ...
package client

import (
	"context"
	"log"
	rumrpc "rum/app/misc/rum"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func POST(addr string, call []*rumrpc.IPost) error {

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}
	defer conn.Close()

	client := rumrpc.NewOnRumServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	x := rumrpc.IPostRequest{
		Post: call,
	}
	if res, err := client.POST(ctx, &x); err != nil {
		return err
	} else {
		log.Println("req succeed: ", res.Succeed)
	}
	return nil
}

func DELETEPROFILE(addr string, call []*rumrpc.IDelete) error {

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}
	defer conn.Close()

	client := rumrpc.NewOnRumServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	x := rumrpc.IDeleteRequest{
		Delete: call,
	}
	if res, err := client.DELETE(ctx, &x); err != nil {
		return err
	} else {
		log.Println("req succeed: ", res.Succeed)
	}
	return nil
}

func ACTIVATEPROFILE(addr string, call []*rumrpc.IActivate) error {

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}
	defer conn.Close()

	client := rumrpc.NewOnRumServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	x := rumrpc.IActivateRequest{
		Activate: call,
	}
	if res, err := client.ACTIVATE(ctx, &x); err != nil {
		return err
	} else {
		log.Println("req succeed: ", res.Succeed)
	}
	return nil
}
func DEACTIVATEPROFILE(addr string, call []*rumrpc.IDeactivate) error {

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}
	defer conn.Close()

	client := rumrpc.NewOnRumServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	x := rumrpc.IDeactivateRequest{
		Deactivate: call,
	}
	if res, err := client.DEACTIVATE(ctx, &x); err != nil {
		return err
	} else {
		log.Println("req succeed: ", res.Succeed)
	}
	return nil
}

func REMOVESERVICE(addr string, call []*rumrpc.IDelete) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := rumrpc.NewOnRumServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	x := rumrpc.IRemoveServiceRequest{Delete: call}
	if res, err := client.REMOVESERVICE(ctx, &x); err != nil {
		return err
	} else {
		log.Println("req succeed: ", res.Succeed)
	}
	return nil
}

func DEACTIVATESERVICE(addr string, call []*rumrpc.IDelete) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := rumrpc.NewOnRumServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	x := rumrpc.IDeactivateServiceRequest{Delete: call}
	if res, err := client.DEACTIVATESERVICE(ctx, &x); err != nil {
		return err
	} else {
		log.Println("req succeed: ", res.Succeed)
	}
	return nil
}

func ACTIVATESERVICE(addr string, call []*rumrpc.IDelete) error {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := rumrpc.NewOnRumServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	x := rumrpc.IActivateServiceRequest{Delete: call}
	if res, err := client.ACTIVATESERVICE(ctx, &x); err != nil {
		return err
	} else {
		log.Println("req succeed: ", res.Succeed)
	}
	return nil
}
