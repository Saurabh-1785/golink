package proto

import (
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/encoding"
)

// jsonCodec implements the gRPC encoding.Codec interface using JSON.
// This lets us use plain Go structs with gRPC instead of protobuf.
type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return "proto" // must be "proto" to be the default codec
}

func (jsonCodec) String() string {
	return fmt.Sprintf("json-as-proto codec")
}

func init() {
	// Register our JSON codec as the default "proto" codec.
	// This means gRPC will use JSON instead of protobuf encoding.
	encoding.RegisterCodec(jsonCodec{})
}
