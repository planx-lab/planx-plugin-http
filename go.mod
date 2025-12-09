module github.com/planx-lab/planx-plugin-http

go 1.25.3

replace github.com/planx-lab/planx-common => ../planx-common

replace github.com/planx-lab/planx-proto => ../planx-proto

replace github.com/planx-lab/planx-sdk-go => ../planx-sdk-go

require (
	github.com/planx-lab/planx-common v0.0.0-00010101000000-000000000000
	github.com/planx-lab/planx-proto v0.0.0-00010101000000-000000000000
	github.com/planx-lab/planx-sdk-go v0.0.0-00010101000000-000000000000
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/grpc v1.77.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
