package client

import (
	"bytes"
	"fmt"
	"github.com/kr/binarydist"
	"github.com/pkg/errors"
	"github.com/wangkechun/vv/pkg/header"
	pb "github.com/wangkechun/vv/pkg/proto"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"qiniupkg.com/x/log.v7"
	"time"
)

const maxEditFileSize = 20 * 1024 * 1024

// Client 组件
type Client struct {
	cfg  Config
	stat os.FileInfo
	args pb.OpenFileRequest
}

// Config Client 配置
type Config struct {
	Name            string
	FilePath        string
	RegistryAddrTCP string
	RegistryAddrRPC string
}

// New 返回一个Client
func New(cfg Config) *Client {
	r := &Client{cfg: cfg}
	return r
}

// Run 启动
func (r *Client) Run() (err error) {
	{
		f := r.cfg.FilePath
		r.args.FileName = filepath.Base(f)
		r.args.Dir = filepath.Dir(f)
		r.stat, err = os.Stat(f)
		if os.IsNotExist(err) {
			log.Info("create", f)
			var file *os.File
			file, err = os.Create(f)
			if err != nil {
				return errors.Wrap(err, "create file failed")
			}
			file.Close()
			r.stat, err = os.Stat(f)
		}
		if err != nil {
			return errors.Wrap(err, "stat file failed")
		}
		if r.stat.Size() > maxEditFileSize {
			return errors.Errorf("file size > %dMB", maxEditFileSize/1024/1024)
		}
		r.args.Content, err = ioutil.ReadFile(f)
		if err != nil {
			return errors.Wrap(err, "open file failed")
		}
	}
	log.Info("dial", r.cfg.RegistryAddrTCP, r.cfg.Name)
	conn, err := net.Dial("tcp", r.cfg.RegistryAddrTCP)
	if err != nil {
		return errors.Wrap(err, "failed to connect registry")
	}
	err = header.WriteHeader(conn, &pb.ProtoHeader{
		Version:    "1",
		User:       r.cfg.Name,
		ServerKind: pb.ProtoHeader_CLIENT,
		ConnKind:   pb.ProtoHeader_DIAL,
	})
	if err != nil {
		return errors.Wrap(err, "failed to connect registry: write header")
	}
	ctx := context.Background()
	var client pb.VvServerClient
	for i := 0; i < 5; i++ {
		err := func() error {
			gconn, err := grpc.Dial("", grpc.WithInsecure(), grpc.WithDialer(func(string, time.Duration) (net.Conn, error) {
				return conn, nil
			}))
			client = pb.NewVvServerClient(gconn)
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			pingReply, err := client.Ping(ctx, &pb.PingRequest{Name: r.cfg.Name})
			if err != nil {
				return errors.Wrap(err, "server reply")
			}
			fmt.Println("file will open in ", pingReply.Name)
			return nil
		}()
		if err == nil {
			break
		}
		if grpc.Code(errors.Cause(err)) == codes.Unavailable {
			continue
		}
		return err
	}

	fileClient, err := client.OpenFile(ctx, &r.args)
	if err != nil {
		return errors.Wrap(err, "server reply")
	}
	for {
		openFileReply, err := fileClient.Recv()
		if err == io.EOF {
			break
		}
		if grpc.Code(err) == codes.Aborted {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "server recv error")
		}
		err = r.applyOpenFileReply(openFileReply)
		if err != nil {
			return errors.Wrap(err, "apply patch error")
		}
	}
	return nil
}

func (r *Client) applyOpenFileReply(reply *pb.OpenFileReply) (err error) {
	if !reply.IsBsdiff {
		return ioutil.WriteFile(r.cfg.FilePath, reply.Content, r.stat.Mode())
	}
	newFile := &bytes.Buffer{}
	err = binarydist.Patch(bytes.NewReader(r.args.Content), newFile, bytes.NewReader(reply.Content))
	if err != nil {
		return errors.Wrap(err, "patch error")
	}
	return ioutil.WriteFile(r.cfg.FilePath, reply.Content, r.stat.Mode())
}
