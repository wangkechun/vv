package header

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	pb "github.com/wangkechun/vv/proto"
	"net"
)

const maxHeaderSize = 1024

func ReadHeader(conn net.Conn) (h *pb.ProtoHeader, err error) {
	var dataLength uint32
	err = binary.Read(conn, binary.BigEndian, &dataLength)
	if err != nil {
		return nil, errors.Wrap(err, "readHeader: read length from conn")
	}
	if dataLength > maxHeaderSize {
		return nil, errors.New("readHeader: too large length")
	}
	buf := make([]byte, dataLength)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "readHeader: read ProtoHeader from conn: read error")
	}
	if n != int(dataLength) {
		return nil, errors.New("readHeader: read ProtoHeader from conn: bad length")
	}
	h = &pb.ProtoHeader{}
	err = proto.Unmarshal(buf, h)
	if err != nil {
		return nil, errors.Wrap(err, "readHeader: read ProtoHeader from conn: decode error")
	}
	return h, nil
}

func WriteHeader(conn net.Conn, h *pb.ProtoHeader) (err error) {
	buf, err := proto.Marshal(h)
	if err != nil {
		return errors.Wrap(err, "writeHeader: proto.Marshal")
	}
	var l uint32 = uint32(len(buf))
	err = binary.Write(conn, binary.BigEndian, l)
	if err != nil {
		return errors.Wrap(err, "writeHeader: binary.Write")
	}
	n, err := conn.Write(buf)
	if err != nil {
		return errors.Wrap(err, "writeHeader: conn.Write")
	}
	if n != len(buf) {
		return errors.Wrap(err, "writeHeader: conn.Write : bad length")
	}
	return nil
}
