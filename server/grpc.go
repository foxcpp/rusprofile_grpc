package server

import (
	"context"
	"log"
	"os"

	"github.com/dgraph-io/ristretto"
	"github.com/foxcpp/rusprofile_grpc/gen/proto"
	"github.com/foxcpp/rusprofile_grpc/parser"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct{
	protodef.UnimplementedRusprofileServer
	cache *ristretto.Cache
	log *log.Logger
}

func New() *Server {
	c, err := ristretto.NewCache(&ristretto.Config{
		NumCounters:        1000,
		MaxCost:            10000,
		BufferItems:        16,
	})
	if err != nil {
		panic(err)
	}
	return &Server{
		cache: c,
		log: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (r Server) Query(ctx context.Context, query *protodef.QueryData) (*protodef.QueryResult, error) {
	var info *parser.CompanyInfo

	val, ok := r.cache.Get(query.Inn)
	if ok {
		info = val.(*parser.CompanyInfo)
	} else {
		var err error
		info, err = parser.Query(ctx, query.Inn)
		if err != nil {
			r.log.Println("Query for", query.Inn, "failed:", err)
			return nil, status.Error(codes.Internal, "internal server error")
		}
		r.cache.Set(query.Inn, info, 1)
	}

	if info == nil {
		return nil, status.Error(codes.NotFound, "no such company")
	}

	res := &protodef.QueryResult{
		Inn: query.Inn,
		DirectorName: info.DirectorName,
		Title: info.Title,
	}
	if info.KPP != "" {
		res.Kpp = &info.KPP
	}
	return res, nil
}
