package fusefs

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/infinitime/blefs"
	"go.arsenm.dev/logger/log"
)

type ITProperty struct {
	name string
	Ino  uint64
	gen  func() ([]byte, error)
}

type DirEntry struct {
	isDir   bool
	modtime uint64
	size    uint32
	path    string
}

type ITNode struct {
	fs.Inode
	kind nodeKind
	Ino  uint64

	lst  []DirEntry
	self DirEntry
	path string
}

type nodeKind uint8

const (
	nodeKindRoot = iota
	nodeKindInfo
	nodeKindFS
	nodeKindReadOnly
)

var (
	myfs     *blefs.FS         = nil
	inodemap map[string]uint64 = nil
)

func BuildRootNode(dev *infinitime.Device) (*ITNode, error) {
	var err error
	inodemap = make(map[string]uint64)
	myfs, err = dev.FS()
	if err != nil {
		log.Error("FUSE Failed to get filesystem").Err(err).Send()
		return nil, err
	}

	return &ITNode{kind: nodeKindRoot}, nil
}

var properties = make([]ITProperty, 6)

func BuildProperties(dev *infinitime.Device) {
	properties[0] = ITProperty{
		"heartrate", 2,
		func() ([]byte, error) {
			ans, err := dev.HeartRate()
			return []byte(strconv.Itoa(int(ans)) + "\n"), err
		},
	}
	properties[1] = ITProperty{
		"battery", 3,
		func() ([]byte, error) {
			ans, err := dev.BatteryLevel()
			return []byte(strconv.Itoa(int(ans)) + "\n"), err
		},
	}
	properties[2] = ITProperty{
		"motion", 4,
		func() ([]byte, error) {
			ans, err := dev.Motion()
			return []byte(strconv.Itoa(int(ans.X)) + " " + strconv.Itoa(int(ans.Y)) + " " + strconv.Itoa(int(ans.Z)) + "\n"), err
		},
	}
	properties[3] = ITProperty{
		"stepcount", 6,
		func() ([]byte, error) {
			ans, err := dev.StepCount()
			return []byte(strconv.Itoa(int(ans)) + "\n"), err
		},
	}
	properties[4] = ITProperty{
		"version", 7,
		func() ([]byte, error) {
			ans, err := dev.Version()
			return []byte(ans + "\n"), err
		},
	}
	properties[5] = ITProperty{
		"address", 8,
		func() ([]byte, error) {
			ans := dev.Address()
			return []byte(ans + "\n"), nil
		},
	}
}

var _ fs.NodeReaddirer = (*ITNode)(nil)

// Readdir is part of the NodeReaddirer interface
func (n *ITNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	switch n.kind {
	case 0:
		// root folder
		r := make([]fuse.DirEntry, 2)
		r[0] = fuse.DirEntry{
			Name: "info",
			Ino:  0,
			Mode: fuse.S_IFDIR,
		}
		r[1] = fuse.DirEntry{
			Name: "fs",
			Ino:  1,
			Mode: fuse.S_IFDIR,
		}
		return fs.NewListDirStream(r), 0

	case 1:
		// info folder
		r := make([]fuse.DirEntry, 6)
		for ind, value := range properties {
			r[ind] = fuse.DirEntry{
				Name: value.name,
				Ino:  value.Ino,
				Mode: fuse.S_IFREG,
			}
		}

		return fs.NewListDirStream(r), 0

	case 2:
		// on info
		files, err := myfs.ReadDir(n.path)
		if err != nil {
			log.Error("FUSE ReadDir failed").Str("path", n.path).Err(err).Send()
			return nil, syscallErr(err)
		}

		log.Debug("FUSE ReadDir succeeded").Str("path", n.path).Int("objects", len(files)).Send()
		r := make([]fuse.DirEntry, len(files))
		n.lst = make([]DirEntry, len(files))
		for ind, entry := range files {
			info, err := entry.Info()
			if err != nil {
				log.Error("FUSE Info failed").Str("path", n.path).Err(err).Send()
				return nil, syscallErr(err)
			}
			name := info.Name()

			file := DirEntry{
				path:    n.path + "/" + name,
				size:    uint32(info.Size()),
				modtime: uint64(info.ModTime().Unix()),
				isDir:   info.IsDir(),
			}
			n.lst[ind] = file

			ino := inodemap[file.path]
			if ino == 0 {
				ino = uint64(len(inodemap)) + 1
				inodemap[file.path] = ino
			}

			if file.isDir {
				r[ind] = fuse.DirEntry{
					Name: name,
					Mode: fuse.S_IFDIR,
					Ino:  ino + 10,
				}
			} else {
				r[ind] = fuse.DirEntry{
					Name: name,
					Mode: fuse.S_IFREG,
					Ino:  ino + 10,
				}
			}
		}
		return fs.NewListDirStream(r), 0
	}
	r := make([]fuse.DirEntry, 0)
	return fs.NewListDirStream(r), 0
}

var _ fs.NodeLookuper = (*ITNode)(nil)

func (n *ITNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	switch n.kind {
	case 0:
		// root folder
		if name == "info" {
			stable := fs.StableAttr{
				Mode: fuse.S_IFDIR,
				Ino:  uint64(0),
			}
			operations := &ITNode{kind: nodeKindInfo, Ino: 0}
			child := n.NewInode(ctx, operations, stable)
			return child, 0
		} else if name == "fs" {
			stable := fs.StableAttr{
				Mode: fuse.S_IFDIR,
				Ino:  uint64(1),
			}
			operations := &ITNode{kind: nodeKindFS, Ino: 1, path: ""}
			child := n.NewInode(ctx, operations, stable)
			return child, 0
		}
	case 1:
		// info folder
		for _, value := range properties {
			if value.name == name {
				stable := fs.StableAttr{
					Mode: fuse.S_IFREG,
					Ino:  uint64(value.Ino),
				}
				operations := &ITNode{kind: nodeKindReadOnly, Ino: value.Ino}
				child := n.NewInode(ctx, operations, stable)
				return child, 0
			}
		}

	case 2:
		// FS object
		if len(n.lst) == 0 {
			n.Readdir(ctx)
		}

		for _, file := range n.lst {
			if file.path != n.path+"/"+name {
				continue
			}
			log.Debug("FUSE Lookup successful").Str("path", file.path).Send()

			if file.isDir {
				stable := fs.StableAttr{
					Mode: fuse.S_IFDIR,
					Ino:  inodemap[file.path],
				}
				operations := &ITNode{kind: nodeKindFS, path: file.path}
				child := n.NewInode(ctx, operations, stable)
				return child, 0
			} else {
				stable := fs.StableAttr{
					Mode: fuse.S_IFREG,
					Ino:  inodemap[file.path],
				}
				operations := &ITNode{
					kind: nodeKindFS, path: file.path,
					self: file,
				}
				child := n.NewInode(ctx, operations, stable)
				return child, 0
			}
			break
		}
		log.Warn("FUSE Lookup failed").Str("path", n.path+"/"+name).Send()
	}
	return nil, syscall.ENOENT
}

type bytesFileReadHandle struct {
	content []byte
}

var _ fs.FileReader = (*bytesFileReadHandle)(nil)

func (fh *bytesFileReadHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.Debug("FUSE Executing Read").Int("size", len(fh.content)).Send()
	end := off + int64(len(dest))
	if end > int64(len(fh.content)) {
		end = int64(len(fh.content))
	}
	return fuse.ReadResultData(fh.content[off:end]), 0
}

type sensorFileReadHandle struct {
	content []byte
}

var _ fs.FileReader = (*sensorFileReadHandle)(nil)

func (fh *sensorFileReadHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	log.Info("Executing Read").Int("size", len(fh.content)).Send()
	end := off + int64(len(dest))
	if end > int64(len(fh.content)) {
		end = int64(len(fh.content))
	}
	return fuse.ReadResultData(fh.content[off:end]), 0
}

var _ fs.FileFlusher = (*sensorFileReadHandle)(nil)

func (fh *sensorFileReadHandle) Flush(ctx context.Context) (errno syscall.Errno) {
	return 0
}

type bytesFileWriteHandle struct {
	content []byte
	path    string
}

var _ fs.FileWriter = (*bytesFileWriteHandle)(nil)

func (fh *bytesFileWriteHandle) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	log.Info("Executing Write").Str("path", fh.path).Int("prev_size", len(fh.content)).Int("next_size", len(data)).Send()
	if off != int64(len(fh.content)) {
		log.Error("FUSE Write file size changed unexpectedly").Int("expect", int(off)).Int("received", len(fh.content)).Send()
		return 0, syscall.ENXIO
	}
	fh.content = append(fh.content[:], data[:]...)
	return uint32(len(data)), 0
}

var _ fs.FileFlusher = (*bytesFileWriteHandle)(nil)

func (fh *bytesFileWriteHandle) Flush(ctx context.Context) (errno syscall.Errno) {
	log.Debug("FUSE Attempting flush").Str("path", fh.path).Send()
	fp, err := myfs.Create(fh.path, uint32(len(fh.content)))
	if err != nil {
		log.Error("FUSE Flush failed: create").Str("path", fh.path).Err(err).Send()
		return syscallErr(err)
	}

	if len(fh.content) == 0 {
		log.Debug("FUSE Flush no data to write").Str("path", fh.path).Send()
		err = fp.Close()
		if err != nil {
			log.Error("FUSE Flush failed during close").Str("path", fh.path).Err(err).Send()
			return syscallErr(err)
		}
		return 0
	}

	go func() {
		// For every progress event
		for sent := range fp.Progress() {
			log.Debug("FUSE Flush progress").Int("bytes", int(sent)).Int("total", len(fh.content)).Send()
		}
	}()

	r := bytes.NewReader(fh.content)
	nread, err := io.Copy(fp, r)
	if err != nil {
		log.Error("FUSE Flush failed during write").Str("path", fh.path).Err(err).Send()
		fp.Close()
		return syscallErr(err)
	}
	if int(nread) != len(fh.content) {
		log.Error("FUSE Flush failed during write").Str("path", fh.path).Int("expect", len(fh.content)).Int("got", int(nread)).Send()
		fp.Close()
		return syscall.EIO
	}
	err = fp.Close()
	if err != nil {
		log.Error("FUSE Flush failed during close").Str("path", fh.path).Err(err).Send()
		return syscallErr(err)
	}
	log.Debug("FUSE Flush done").Str("path", fh.path).Int("size", len(fh.content)).Send()

	return 0
}

var _ fs.FileFsyncer = (*bytesFileWriteHandle)(nil)

func (fh *bytesFileWriteHandle) Fsync(ctx context.Context, flags uint32) (errno syscall.Errno) {
	return fh.Flush(ctx)
}

var _ fs.NodeGetattrer = (*ITNode)(nil)

func (bn *ITNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	log.Debug("FUSE getattr").Str("path", bn.path).Send()
	out.Ino = bn.Ino
	out.Mtime = bn.self.modtime
	out.Ctime = bn.self.modtime
	out.Atime = bn.self.modtime
	out.Size = uint64(bn.self.size)
	return 0
}

var _ fs.NodeSetattrer = (*ITNode)(nil)

func (bn *ITNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	log.Debug("FUSE setattr").Str("path", bn.path).Send()
	out.Size = 0
	out.Mtime = 0
	return 0
}

var _ fs.NodeOpener = (*ITNode)(nil)

func (f *ITNode) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	switch f.kind {
	case 2:
		// FS file
		if openFlags&syscall.O_RDWR != 0 {
			log.Error("FUSE Open failed: RDWR").Str("path", f.path).Send()
			return nil, 0, syscall.EROFS
		}

		if openFlags&syscall.O_WRONLY != 0 {
			log.Debug("FUSE Opening for write").Str("path", f.path).Send()
			fh = &bytesFileWriteHandle{
				path:    f.path,
				content: make([]byte, 0),
			}
			return fh, fuse.FOPEN_DIRECT_IO, 0
		} else {
			log.Debug("FUSE Opening for read").Str("path", f.path).Send()
			fp, err := myfs.Open(f.path)
			if err != nil {
				log.Error("FUSE: Opening failed").Str("path", f.path).Err(err).Send()
				return nil, 0, syscallErr(err)
			}

			defer fp.Close()

			b := &bytes.Buffer{}

			go func() {
				// For every progress event
				for sent := range fp.Progress() {
					log.Debug("FUSE Read progress").Int("bytes", int(sent)).Int("total", int(f.self.size)).Send()
				}
			}()

			_, err = io.Copy(b, fp)
			if err != nil {
				log.Error("FUSE Read failed").Str("path", f.path).Err(err).Send()
				fp.Close()
				return nil, 0, syscallErr(err)
			}

			fh = &bytesFileReadHandle{
				content: b.Bytes(),
			}
			return fh, fuse.FOPEN_DIRECT_IO, 0
		}

	case 3:
		// Device file

		// disallow writes
		if openFlags&(syscall.O_RDWR|syscall.O_WRONLY) != 0 {
			return nil, 0, syscall.EROFS
		}

		for _, value := range properties {
			if value.Ino == f.Ino {
				ans, err := value.gen()
				if err != nil {
					return nil, 0, syscallErr(err)
				}

				fh = &sensorFileReadHandle{
					content: ans,
				}
				return fh, fuse.FOPEN_DIRECT_IO, 0
			}
		}
	}
	return nil, 0, syscall.EINVAL
}

var _ fs.NodeCreater = (*ITNode)(nil)

func (f *ITNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if f.kind != 2 {
		return nil, nil, 0, syscall.EROFS
	}

	path := f.path + "/" + name
	ino := uint64(len(inodemap)) + 11
	inodemap[path] = ino

	stable := fs.StableAttr{
		Mode: fuse.S_IFREG,
		Ino:  ino,
	}
	operations := &ITNode{
		kind: nodeKindFS, Ino: ino,
		path: path,
	}
	node = f.NewInode(ctx, operations, stable)

	fh = &bytesFileWriteHandle{
		path:    path,
		content: make([]byte, 0),
	}

	log.Debug("FUSE Creating file").Str("path", path).Send()

	errno = 0
	return node, fh, fuseFlags, 0
}

var _ fs.NodeMkdirer = (*ITNode)(nil)

func (f *ITNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if f.kind != 2 {
		return nil, syscall.EROFS
	}

	path := f.path + "/" + name
	err := myfs.Mkdir(path)
	if err != nil {
		log.Error("FUSE Mkdir failed").
			Str("path", path).
			Err(err).
			Send()
		return nil, syscallErr(err)
	}

	ino := uint64(len(inodemap)) + 11
	inodemap[path] = ino

	stable := fs.StableAttr{
		Mode: fuse.S_IFDIR,
		Ino:  ino,
	}
	operations := &ITNode{
		kind: nodeKindFS, Ino: ino,
		path: path,
	}
	node := f.NewInode(ctx, operations, stable)

	log.Debug("FUSE Mkdir success").
		Str("path", path).
		Int("ino", int(ino)).
		Send()
	return node, 0
}

var _ fs.NodeRenamer = (*ITNode)(nil)

func (f *ITNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	if f.kind != 2 {
		return syscall.EROFS
	}

	p1 := f.path + "/" + name
	p2 := newParent.EmbeddedInode().Path(nil)[2:] + "/" + newName

	err := myfs.Rename(p1, p2)
	if err != nil {
		log.Error("FUSE Rename failed").
			Str("src", p1).
			Str("dest", p2).
			Err(err).
			Send()

		return syscallErr(err)
	}
	log.Debug("FUSE Rename sucess").
		Str("src", p1).
		Str("dest", p2).
		Send()

	ino := inodemap[p1]
	delete(inodemap, p1)
	inodemap[p2] = ino

	return 0
}

var _ fs.NodeUnlinker = (*ITNode)(nil)

func (f *ITNode) Unlink(ctx context.Context, name string) syscall.Errno {
	if f.kind != 2 {
		return syscall.EROFS
	}

	delete(inodemap, f.path+"/"+name)
	err := myfs.Remove(f.path + "/" + name)
	if err != nil {
		log.Error("FUSE Unlink failed").
			Str("file", f.path+"/"+name).
			Err(err).
			Send()

		return syscallErr(err)
	}

	log.Debug("FUSE Unlink success").
		Str("file", f.path+"/"+name).
		Send()
	return 0
}

var _ fs.NodeRmdirer = (*ITNode)(nil)

func (f *ITNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	return f.Unlink(ctx, name)
}
