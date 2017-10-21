package server

import (
	"bytes"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/wangkechun/vv/editor"
	pb "github.com/wangkechun/vv/proto"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"qiniupkg.com/x/log.v7"
)

const defaultConnectionNum = 5

// server is used to implement helloworld.GreeterServer.
type fileServer struct {
	name string
}

// SayHello implements helloworld.GreeterServer
func (s *fileServer) Ping(ctx context.Context, in *pb.PingRequest) (out *pb.PingReply, err error) {
	return &pb.PingReply{Name: s.name}, nil
}

func (s *fileServer) OpenFile(in *pb.OpenFileRequest, stream pb.VvServer_OpenFileServer) (err error) {
	// TODO:更好的处理重名问题
	fileName := filepath.Join(os.TempDir(), in.FileName)
	err = ioutil.WriteFile(fileName, in.Content, 0600)
	if err != nil {
		return errors.New("write file")
	}
	defer os.Remove(fileName)

	log.Info("write file", fileName)
	command := editor.Cmd(fileName)
	log.Info("run", command)

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	err = cmd.Start()

	if err != nil {
		return errors.Wrap(err, "open editor error")
	}

	go func() {
		cmd.Wait()
		log.Info("close editor")
		cancel()
	}()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "fsnotify.NewWatcher")
	}
	watcher.Add(fileName)
	defer watcher.Close()

	lastContnet := in.Content

	for {
		select {
		case <-ctx.Done():
			return grpc.Errorf(codes.Aborted, "editor or client closed")
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)
				// TODO: bsdiff 支持
				buf, err := ioutil.ReadFile(fileName)
				if err != nil {
					return errors.Wrap(err, "read file error")
				}
				if bytes.Compare(lastContnet, buf) == 0 {
					continue
				}
				lastContnet = buf
				log.Infof("server, file %s change , push it", fileName)
				err = stream.Send(&pb.OpenFileReply{
					Content: lastContnet,
				})
				if err != nil {
					log.Error("send file error", err)
					return err
				}
			}
		case err := <-watcher.Errors:
			log.Error("watch file error", err)
			return err
		}
	}
}

// Server 组件
type Server struct {
	cfg Config
}

// Config Server 配置
type Config struct {
	RegistryAddr string
	Name         string
}

// New 返回一个Server
func New(cfg Config) *Server {
	r := &Server{cfg: cfg}
	return r
}

// Run 启动
func (r *Server) Run() (err error) {
	srv := &fileServer{
		name: r.cfg.Name,
	}
	s := grpc.NewServer()
	pb.RegisterVvServerServer(s, srv)
	reflection.Register(s)
	pool := newPool(r.cfg, defaultConnectionNum)

	err = s.Serve(&pool)
	if err != nil {
		return errors.Wrap(err, "failed to connect registry: serve")
	}
	return nil
}
