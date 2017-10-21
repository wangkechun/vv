package registry

import (
	"github.com/stretchr/testify/assert"
	pb "github.com/wangkechun/vv/proto"
	"net"
	"testing"
)

func TestReadWriteHeader(t *testing.T) {
	assert := assert.New(t)
	server, client := net.Pipe()
	go writeHeader(client, &pb.ProtoHeader{
		Version:    "1",
		Token:      "123",
		ServerKind: pb.ProtoHeader_CLIENT,
		ConnKind:   pb.ProtoHeader_DIAL,
	})
	header, err := readHeader(server)
	assert.Nil(err)
	assert.Equal(header.Token, "123")
}
