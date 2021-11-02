package main

import (
	"context"
	"fmt"
	"os"

	protodef "github.com/foxcpp/rusprofile_grpc/gen/proto"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s <server endpoint> <inn>\n", os.Args[0])
		os.Exit(2)
	}

	conn, err := grpc.Dial(os.Args[1], grpc.WithInsecure())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer conn.Close()

	client := protodef.NewRusprofileClient(conn)
	res, err := client.Query(context.Background(), &protodef.QueryData{Inn: os.Args[2]})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(res)
}
