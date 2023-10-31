protoc --proto_path=proto proto/user.proto --go_out=plugins=grpc:.
protoc --proto_path=proto proto/rpc_create_user.proto --go_out=plugins=grpc:.
protoc --proto_path=proto proto/rpc_update_user.proto --go_out=plugins=grpc:.
protoc --proto_path=proto proto/rpc_login_user.proto --go_out=plugins=grpc:.
protoc --proto_path=proto proto/service_simple_bank.proto --go_out=plugins=grpc:. --grpc-gateway_out=:. --openapiv2_out=:docs/swagger