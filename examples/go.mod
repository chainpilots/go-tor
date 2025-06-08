module github.com/chainpilots/go-tor/examples

go 1.24.3

require (
	github.com/chainpilots/go-tor v0.2.0
	github.com/golang/protobuf v1.5.4
	golang.org/x/net v0.41.0
	google.golang.org/grpc v1.73.0
)

require (
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/chainpilots/go-tor => ../
