package lab

import (
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/monochromegane/go-gitignore"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	myfuse "github.com/wangkechun/vv/pkg/fuse"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"qiniupkg.com/x/log.v7"
	"syscall"
	"testing"
	"time"
)

func getInitFs(dir string) (dirs []string, files []string, err error) {
	gitignores := make([]gitignore.IgnoreMatcher, 0)
	{
		d := dir
		for i := 0; i < 10; i++ {
			gitignore, err := gitignore.NewGitIgnore(path.Join(d, ".gitignore"))
			if err == nil {
				fmt.Println("find .gitignore", path.Join(d, ".gitignore"))
				gitignores = append(gitignores, gitignore)
			}
			d2 := path.Dir(d)
			if d2 == d {
				break
			}
			d = d2
		}
	}
	n := 0

	dirs = make([]string, 0)
	files = make([]string, 0)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		n++
		if n > 5000 {
			return errors.New("too many file")
		}

		isDir := info.IsDir()
		filePath := path
		if isDir {
			filePath = filepath.Join(path, info.Name())
		}
		// log.Info("walk", filePath)

		for _, gitignore := range gitignores {
			if gitignore.Match(filePath, isDir) {
				return filepath.SkipDir
			}
		}
		relDir, err := filepath.Rel(dir, filePath)
		if isDir {
			dirs = append(dirs, relDir)
		} else {
			files = append(files, relDir)
		}
		return nil
	})
	return dirs, files, nil
}

func TestGitignore(t *testing.T) {
	a := assert.New(t)
	dir := "/Users/wkc/qbox/ava/atshow/front"
	dirs, files, err := getInitFs(dir)
	a.Nil(err)
	log.Info(dirs)
	log.Info(files)
}

type netFs struct {
	fileName string
}

func (f *netFs) Stat(out *fuse.Attr) {
	stat, err := os.Stat(filepath.Join("/Users/wkc/qbox/ava/atshow/front", f.fileName))
	if err != nil {
		log.Error(f, err)
	}
	out.Size = uint64(stat.Size())
	out.Mode = uint32(stat.Mode()) | 0666
	// log.Infof("%o %o\n", out.Mode, out.Mode|0666)

	t := stat.ModTime()
	out.SetTimes(&t, &t, &t)
	out.Mode |= syscall.S_IFREG
}

func (f *netFs) Data() []byte {
	buf, err := ioutil.ReadFile(filepath.Join("/Users/wkc/qbox/ava/atshow/front", f.fileName))
	if err != nil {
		log.Error(f, err)
	}
	return buf
}

func (f *netFs) Update(data []byte) {
	err := ioutil.WriteFile(filepath.Join("/Users/wkc/qbox/ava/atshow/front", f.fileName), data, 0644)
	if err != nil {
		log.Error(f, err)
	}
}

func inc() {
	go http.Get("http://localhost:8855")
}

func FunFuseG1() {
	dir := "/Users/wkc/qbox/ava/atshow/front"
	root := "/Users/wkc/Downloads/tmp"
	_, files, err := getInitFs(dir)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("files num ", len(files))
	fmt.Println("send all files...")
	time.Sleep(time.Second * 3)
	inc()
	fmt.Println("watch change...")
	cmd := exec.Command("tail", "-f", "../1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
}

func FunFuse() {
	dir := "/Users/wkc/qbox/ava/atshow/front"
	root := "/Users/wkc/Downloads/tmp"
	_, files, err := getInitFs(dir)
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("files", len(files))
	memfiles := map[string]myfuse.MemFile{}
	for _, v := range files {
		memfiles[v] = &netFs{
			fileName: v,
		}
	}
	mfs := myfuse.NewMemTreeFs(memfiles)
	mfs.Name = "fs(netfs)"
	opts := &nodefs.Options{
		AttrTimeout:  time.Duration(time.Second),
		EntryTimeout: time.Duration(time.Second),
		Debug:        false,
	}
	server, _, err := nodefs.MountRoot(root, mfs.Root(), opts)
	if err != nil {
		log.Panic(err)
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for sig := range c {
			err := server.Unmount()
			if err != nil {
				log.Info("fuse umount error", err)
				os.Exit(1)
			}
			// 强制umount？
			// umount -f /Users/wkc/Downloads/tmp
			os.Exit(0)
		}
	}()
	server.Serve()
}
