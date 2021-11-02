package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	protodef "github.com/foxcpp/rusprofile_grpc/gen/proto"
	"github.com/foxcpp/rusprofile_grpc/server"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

func main() {
	grpcEndp := flag.String("grpc-addr", "localhost:8888", "gRPC endpoint to listen on")
	httpEndp := flag.String("http-addr", "localhost:8080", "HTTP endpoint to listen on")
	staticLoc := flag.String("static", "./static", "Static files for built-in Swagger UI")
	flag.Parse()

	grpcL, err := net.Listen("tcp", *grpcEndp)
	if err != nil {
		log.Fatalln(err)
	}
	httpL, err := net.Listen("tcp", *httpEndp)
	if err != nil {
		log.Fatalln(err)
	}

	srv := server.New()
	grpcSrv := grpc.NewServer()
	protodef.RegisterRusprofileServer(grpcSrv, srv)
	log.Println("Listening on", *grpcEndp, "for gRPC requests...")
	go grpcSrv.Serve(grpcL)

	mux := http.NewServeMux()
	gwMux := runtime.NewServeMux()
	httpSrv := http.Server{Handler: mux}
	ctx, cancel := context.WithCancel(context.Background())
	protodef.RegisterRusprofileHandlerFromEndpoint(ctx, gwMux, *grpcEndp,
		[]grpc.DialOption{grpc.WithInsecure()})

	mux.Handle("/swagger/", http.StripPrefix("/swagger/", http.FileServer(http.Dir(*staticLoc))))
	mux.Handle("/v1/", gwMux)
	log.Println("Listening on", *httpEndp, "for HTTP requests...")

	go httpSrv.Serve(httpL)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	<-sig
	log.Println("Waiting for existing queries to complete...")
	done := make(chan struct{})
	cancel()
	go func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 26*time.Second)
		defer cancel()

		grpcSrv.GracefulStop()
		httpSrv.Shutdown(stopCtx)
		done <- struct{}{}
	}()
	select {
	case <-sig:
		log.Println("Force stop...")
		grpcSrv.Stop()
	case <-done:
	}
}
