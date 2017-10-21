package fuse

import (
	"fmt"
	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"qiniupkg.com/x/log.v7"
)

// DataFile is for implementing read-only filesystems.  This
// assumes we already have the data in memory.
type dataFile struct {
	m MemFile

	nodefs.File
}

func (f *dataFile) String() string {
	l := len(f.m.Data())
	if l > 10 {
		l = 10
	}

	return fmt.Sprintf("dataFile(%x)", f.m.Data()[:l])
}

func (f *dataFile) GetAttr(out *fuse.Attr) fuse.Status {
	out.Mode = fuse.S_IFREG | 0644
	out.Size = uint64(len(f.m.Data()))
	return fuse.OK
}

func NewDataFile(m MemFile) nodefs.File {
	f := new(dataFile)
	f.m = m
	f.File = NewDefaultFile()
	return f
}

func (f *dataFile) Read(buf []byte, off int64) (res fuse.ReadResult, code fuse.Status) {
	log.Info("read", off, len(buf))
	end := int(off) + int(len(buf))
	if end > len(f.m.Data()) {
		end = len(f.m.Data())
	}

	return fuse.ReadResultData(f.m.Data()[off:end]), fuse.OK
}

func (f *dataFile) Write(data []byte, off int64) (uint32, fuse.Status) {
	log.Info("write", off, len(data))
	old := f.m.Data()
	addSize := int(off) + len(data) - len(old)
	if addSize > 0 {
		buf := make([]byte, int(off)+len(data))
		copy(buf, old)
		old = buf
	}
	n := copy(old[off:], data)
	if n != len(data) {
		panic("")
	}
	f.m.Update(old)
	return uint32(len(data)), fuse.OK
}
