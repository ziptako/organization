```shell
  goctl rpc protoc .\organization.proto -m --style goZero --zrpc_out . --go-grpc_out . --go_out .
```

```shell
  go build -o ./build/main.exe .
```

```shell
  docker compose up --build
```

```shell
  build/main.exe
```