package fuse

import (
	"fmt"
	"strings"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"qiniupkg.com/x/log.v7"
)

type MemFile interface {
	Stat(out *fuse.Attr)
	Data() []byte
	Update([]byte)
}

type memNode struct {
	nodefs.Node
	file MemFile
	fs   *MemTreeFs
}

// memTreeFs creates a tree of internal Inodes.  Since the tree is
// loaded in memory completely at startup, it does not need inode
// discovery through Lookup() at serve time.
type MemTreeFs struct {
	root  *memNode
	files map[string]MemFile
	Name  string
}

func NewMemTreeFs(files map[string]MemFile) *MemTreeFs {
	fs := &MemTreeFs{
		root:  &memNode{Node: nodefs.NewDefaultNode()},
		files: files,
	}
	fs.root.fs = fs
	return fs
}

func (fs *MemTreeFs) String() string {
	return fs.Name
}

func (fs *MemTreeFs) Root() nodefs.Node {
	return fs.root
}

func (fs *MemTreeFs) onMount() {
	for k, v := range fs.files {
		fs.addFile(k, v)
	}
	fs.files = nil
}

func (n *memNode) OnMount(c *nodefs.FileSystemConnector) {
	log.Info("memNode.OnMount")
	n.fs.onMount()
}

func (n *memNode) Print(indent int) {
	log.Info("memNode.Print")
	s := ""
	for i := 0; i < indent; i++ {
		s = s + " "
	}

	children := n.Inode().Children()
	for k, v := range children {
		if v.IsDir() {
			fmt.Println(s + k + ":")
			mn, ok := v.Node().(*memNode)
			if ok {
				mn.Print(indent + 2)
			}
		} else {
			fmt.Println(s + k)
		}
	}
}

func (n *memNode) OpenDir(context *fuse.Context) (stream []fuse.DirEntry, code fuse.Status) {
	log.Info("memNode.OpenDir")
	children := n.Inode().Children()
	stream = make([]fuse.DirEntry, 0, len(children))
	for k, v := range children {
		mode := fuse.S_IFREG | 0666
		if v.IsDir() {
			mode = fuse.S_IFDIR | 0777
		}
		stream = append(stream, fuse.DirEntry{
			Name: k,
			Mode: uint32(mode),
		})
	}
	return stream, fuse.OK
}

func (n *memNode) Open(flags uint32, context *fuse.Context) (fuseFile nodefs.File, code fuse.Status) {
	// if flags&fuse.O_ANYWRITE != 0 {
	// 	return nil, fuse.EPERM
	// }
	log.Infof("memNode.Open %o %s", flags, n.fs.Name)
	return NewDataFile(n.file), fuse.OK
}

func (n *memNode) Write(file nodefs.File, data []byte, off int64, context *fuse.Context) (written uint32, code fuse.Status) {
	log.Infof("memNode.Write %o %s", file, n.fs.Name, off)
	// if file != nil {
	// 	return file.Write(data, off)
	// }

	// return 0, fuse.ENOSYS
	// n.file.Update(data)
	// n.file
	old := n.file.Data()
	log.Info("write", off, len(data))
	addSize := int(off) + len(data) - len(old)
	if addSize > 0 {
		buf := make([]byte, int(off)+len(data))
		copy(buf, old)
		old = buf
	}
	cnt := copy(old[off:], data)
	if cnt != len(data) {
		panic("")
	}
	n.file.Update(old)
	return uint32(len(data)), fuse.OK
}

func (n *memNode) Deletable() bool {
	log.Info("memNode.Deletable")
	return false
}

func (n *memNode) GetAttr(out *fuse.Attr, file nodefs.File, context *fuse.Context) fuse.Status {
	log.Info("memNode.GetAttr")
	if n.Inode().IsDir() {
		out.Mode = fuse.S_IFDIR | 0777
		return fuse.OK
	}
	n.file.Stat(out)
	out.Blocks = (out.Size + 511) / 512
	return fuse.OK
}

func (n *memNode) Truncate(file nodefs.File, size uint64, context *fuse.Context) (code fuse.Status) {
	buf := n.file.Data()
	if int(size) > len(buf) {
		size = uint64(len(buf))
	}
	n.file.Update(buf[:size])
	log.Info("memNode.Truncate")
	return fuse.OK
}

func (n *memNode) GetXAttr(attribute string, context *fuse.Context) (data []byte, code fuse.Status) {
	log.Info("memNode.GetXAttr")
	return nil, fuse.OK
}

func (n *MemTreeFs) addFile(name string, f MemFile) {
	comps := strings.Split(name, "/")

	node := n.root.Inode()
	for i, c := range comps {
		child := node.GetChild(c)
		if child == nil {
			fsnode := &memNode{
				Node: nodefs.NewDefaultNode(),
				fs:   n,
			}
			if i == len(comps)-1 {
				fsnode.file = f
			}

			child = node.NewChild(c, fsnode.file == nil, fsnode)
		}
		node = child
	}
}
