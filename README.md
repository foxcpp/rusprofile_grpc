# rusprofile.ru gRPC wrapper

```shell
$ go build ./cmd/rusprofile_service
$ ./rusprofile_service -grpc-endpoint :8000 -http-endpoint :8001
```
```shell
$ curl http://localhost:8001/v1/query/5017096885
{"inn":"5017096885","kpp":"501701001","title":"ООО \"Строй Комплект\"","directorName":"Дудин Алексей Васильевич"}⏎
$ xdg-open http://localhost:8001/swagger/ # Swagger UI
```

```
$ go build ./cmd/rusprofile_query
$ ./rusprofile_query localhost:8000 5017096885
inn:"5017096885"  kpp:"501701001"  title:"ООО \"Строй Комплект\""  directorName:"Дудин Алексей Васильевич"
```

```shell
# docker build -t rusprofile_grpc .
# docker run -d --rm \
  -e GRPC_ENDPOINT=:8000 -e HTTP_ENDPOINT=:8001 \
  -p 8000:8000 -p 8001:8001 rusprofile_grpc
$ curl http://localhost:8001/v1/query/5017096885
{"inn":"5017096885","kpp":"501701001","title":"ООО \"Строй Комплект\"","directorName":"Дудин Алексей Васильевич"}⏎
$ xdg-open http://localhost:8001/swagger # Swagger UI
```
