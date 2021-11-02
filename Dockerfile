FROM golang:1.16.5-alpine3.14 AS build-env

COPY . /rusprofile_grpc/
WORKDIR /rusprofile_grpc/
RUN go build ./cmd/rusprofile_service

FROM alpine:3.14
COPY --from=build-env /rusprofile_grpc/rusprofile_service /bin/rusprofile_service
COPY --from=build-env /rusprofile_grpc/static /static
WORKDIR /
CMD /bin/rusprofile_service --grpc-addr $GRPC_ENDPOINT --http-addr $HTTP_ENDPOINT