package fusefs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"go.elara.ws/itd/infinitime"
)

type DirEntry struct {
	isDir   bool
	modtime uint64
	size    uint32
	path    string
}

type ITNode struct {
	fs.Inode
	Ino      uint64
	inodemap map[string]uint64

	lst  []DirEntry
	self DirEntry
	path string

	devfs *infinitime.FS
	log   *slog.Logger
}

func BuildRootNode(ilog *slog.Logger, dev *infinitime.Device) (*ITNode, error) {
	return &ITNode{
		devfs:    dev.FS(),
		log:      ilog.With(slog.String("component", "fusefs")),
		inodemap: map[string]uint64{},
	}, nil
}

var _ fs.NodeReaddirer = (*ITNode)(nil)

// Readdir is part of the NodeReaddirer interface
func (n *ITNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	// on info
	files, err := n.devfs.ReadDir(n.path)
	if err != nil {
		n.log.Error("FUSE ReadDir failed", slog.String("path", n.path))
		return nil, syscallErr(err)
	}

	n.log.Debug("FUSE ReadDir succeeded", slog.String("path", n.path), slog.Int("objects", len(files)))
	r := make([]fuse.DirEntry, len(files))
	n.lst = make([]DirEntry, len(files))
	for ind, entry := range files {
		info, err := entry.Info()
		if err != nil {
			n.log.Error("FUSE Info failed", slog.String("path", n.path))
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

		ino := n.inodemap[file.path]
		if ino == 0 {
			ino = uint64(len(n.inodemap)) + 1
			n.inodemap[file.path] = ino
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

var _ fs.NodeLookuper = (*ITNode)(nil)

func (n *ITNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	if len(n.lst) == 0 {
		n.Readdir(ctx)
	}

	for _, file := range n.lst {
		if file.path != n.path+"/"+name {
			continue
		}
		n.log.Debug("FUSE Lookup successful", slog.String("path", file.path))

		if file.isDir {
			stable := fs.StableAttr{
				Mode: fuse.S_IFDIR,
				Ino:  n.inodemap[file.path],
			}
			operations := &ITNode{
				path:     file.path,
				log:      n.log,
				inodemap: n.inodemap,
				devfs:    n.devfs,
			}
			child := n.NewInode(ctx, operations, stable)
			return child, 0
		} else {
			stable := fs.StableAttr{
				Mode: fuse.S_IFREG,
				Ino:  n.inodemap[file.path],
			}
			operations := &ITNode{
				path:     file.path,
				self:     file,
				log:      n.log,
				inodemap: n.inodemap,
				devfs:    n.devfs,
			}
			child := n.NewInode(ctx, operations, stable)
			return child, 0
		}
	}
	n.log.Warn("FUSE Lookup failed", slog.String("path", n.path+"/"+name))
	return nil, syscall.ENOENT
}

type bytesFileReadHandle struct {
	df   *infinitime.File
	node *ITNode
}

var _ fs.FileReader = (*bytesFileReadHandle)(nil)

func (fh *bytesFileReadHandle) Read(ctx context.Context, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {
	fh.node.log.Debug("FUSE Executing Read")
	se := syscall.Errno(0)

	n, err := fh.df.Read(dest)
	if err != nil && !errors.Is(err, io.EOF) {
		se = syscallErr(err)
	}

	return fuse.ReadResultData(dest[:n]), se
}

type bytesFileWriteHandle struct {
	buf  *bytes.Buffer
	path string
	node *ITNode
}

var _ fs.FileWriter = (*bytesFileWriteHandle)(nil)

func (fh *bytesFileWriteHandle) Write(ctx context.Context, data []byte, off int64) (written uint32, errno syscall.Errno) {
	fh.node.log.Debug("FUSE Executing Write",
		slog.String("path", fh.path),
		slog.Int("prev_size", fh.buf.Len()),
		slog.Int("next_size", len(data)),
	)
	if off != int64(fh.buf.Len()) {
		fh.node.log.Debug("FUSE Write file size changed unexpectedly",
			slog.Int("expect", int(off)),
			slog.Int("received", fh.buf.Len()),
		)
		return 0, syscall.ENXIO
	}
	n, err := fh.buf.Write(data)
	return uint32(n), syscallErr(err)
}

var _ fs.FileFlusher = (*bytesFileWriteHandle)(nil)

func (fh *bytesFileWriteHandle) Flush(ctx context.Context) (errno syscall.Errno) {
	fh.node.log.Debug("FUSE Attempting flush", slog.String("path", fh.path))
	fp, err := fh.node.devfs.Create(fh.path, uint32(fh.buf.Len()))
	if err != nil {
		fh.node.log.Error("FUSE Flush failed: create", slog.String("path", fh.path))
		return syscallErr(err)
	}

	if fh.buf.Len() == 0 {
		fh.node.log.Debug("FUSE Flush no data to write", slog.String("path", fh.path))
		err = fp.Close()
		if err != nil {
			fh.node.log.Error("FUSE Flush failed during close", slog.String("path", fh.path))
			return syscallErr(err)
		}
		return 0
	}

	fp.ProgressFunc = func(transferred, total uint32) {
		fh.node.log.Debug("FUSE Read progress", slog.Uint64("bytes", uint64(transferred)), slog.Uint64("total", uint64(total)))
	}

	_, err = io.Copy(fp, fh.buf)
	if err != nil {
		fh.node.log.Error("FUSE Flush failed during write", slog.String("path", fh.path))
		fp.Close()
		return syscallErr(err)
	}
	if fh.buf.Len() != 0 {
		fh.node.log.Error("FUSE Flush failed during write",
			slog.String("path", fh.path),
			slog.Int("expect", 0),
			slog.Int("got", fh.buf.Len()),
		)
		fp.Close()
		return syscall.EIO
	}
	err = fp.Close()
	if err != nil {
		fh.node.log.Error("FUSE Flush failed during close", slog.String("path", fh.path))
		return syscallErr(err)
	}
	fh.node.log.Debug("FUSE Flush done", slog.String("path", fh.path), slog.Int("size", fh.buf.Len()))

	return 0
}

var _ fs.FileFsyncer = (*bytesFileWriteHandle)(nil)

func (fh *bytesFileWriteHandle) Fsync(ctx context.Context, flags uint32) (errno syscall.Errno) {
	return fh.Flush(ctx)
}

var _ fs.NodeGetattrer = (*ITNode)(nil)

func (bn *ITNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	bn.log.Debug("FUSE getattr", slog.String("path", bn.path))
	out.Ino = bn.Ino
	out.Mtime = bn.self.modtime
	out.Ctime = bn.self.modtime
	out.Atime = bn.self.modtime
	out.Size = uint64(bn.self.size)
	return 0
}

var _ fs.NodeSetattrer = (*ITNode)(nil)

func (bn *ITNode) Setattr(ctx context.Context, fh fs.FileHandle, in *fuse.SetAttrIn, out *fuse.AttrOut) syscall.Errno {
	bn.log.Debug("FUSE setattr", slog.String("path", bn.path))
	out.Size = 0
	out.Mtime = 0
	return 0
}

var _ fs.NodeOpener = (*ITNode)(nil)

func (f *ITNode) Open(ctx context.Context, openFlags uint32) (fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	if openFlags&syscall.O_RDWR != 0 {
		f.log.Error("FUSE Open failed: RDWR", slog.String("path", f.path))
		return nil, 0, syscall.EROFS
	}

	if openFlags&syscall.O_WRONLY != 0 {
		f.log.Debug("FUSE Opening for write", slog.String("path", f.path))
		fh = &bytesFileWriteHandle{
			path: f.path,
			buf:  &bytes.Buffer{},
			node: f,
		}
		return fh, fuse.FOPEN_DIRECT_IO, 0
	} else {
		f.log.Debug("FUSE Opening for read", slog.String("path", f.path))
		fp, err := f.devfs.Open(f.path)
		if err != nil {
			f.log.Error("FUSE: Opening failed", slog.String("path", f.path))
			return nil, 0, syscallErr(err)
		}

		fp.ProgressFunc = func(transferred, total uint32) {
			f.log.Debug("FUSE Read progress", slog.Uint64("bytes", uint64(transferred)), slog.Uint64("total", uint64(total)))
		}

		fh = &bytesFileReadHandle{
			node: f,
			df:   fp,
		}
		return fh, fuse.FOPEN_DIRECT_IO, 0
	}
}

var _ fs.NodeCreater = (*ITNode)(nil)

func (f *ITNode) Create(ctx context.Context, name string, flags uint32, mode uint32, out *fuse.EntryOut) (node *fs.Inode, fh fs.FileHandle, fuseFlags uint32, errno syscall.Errno) {
	path := f.path + "/" + name
	ino := uint64(len(f.inodemap)) + 11
	f.inodemap[path] = ino

	stable := fs.StableAttr{
		Mode: fuse.S_IFREG,
		Ino:  ino,
	}
	operations := &ITNode{
		Ino:      ino,
		path:     path,
		log:      f.log,
		inodemap: f.inodemap,
		devfs:    f.devfs,
	}
	node = f.NewInode(ctx, operations, stable)

	fh = &bytesFileWriteHandle{
		path: path,
		buf:  &bytes.Buffer{},
		node: f,
	}

	f.log.Debug("FUSE Creating file", slog.String("path", path))

	errno = 0
	return node, fh, fuseFlags, 0
}

var _ fs.NodeMkdirer = (*ITNode)(nil)

func (f *ITNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	path := f.path + "/" + name
	err := f.devfs.Mkdir(path)
	if err != nil {
		f.log.Error("FUSE Mkdir failed", slog.String("path", path), slog.Any("error", err))
		return nil, syscallErr(err)
	}

	ino := uint64(len(f.inodemap)) + 11
	f.inodemap[path] = ino

	stable := fs.StableAttr{
		Mode: fuse.S_IFDIR,
		Ino:  ino,
	}
	operations := &ITNode{
		Ino:      ino,
		path:     path,
		log:      f.log,
		inodemap: f.inodemap,
		devfs:    f.devfs,
	}
	node := f.NewInode(ctx, operations, stable)

	f.log.Debug("FUSE Mkdir success", slog.String("path", path), slog.Int("ino", int(ino)))
	return node, 0
}

var _ fs.NodeRenamer = (*ITNode)(nil)

func (f *ITNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {
	p1 := f.path + "/" + name
	p2 := newParent.EmbeddedInode().Path(nil)[2:] + "/" + newName

	err := f.devfs.Rename(p1, p2)
	if err != nil {
		f.log.Error("FUSE rename failed", slog.String("src", p1), slog.String("dest", p2), slog.Any("error", err))
		return syscallErr(err)
	}
	f.log.Debug("FUSE rename success", slog.String("src", p1), slog.String("dest", p2))

	ino := f.inodemap[p1]
	delete(f.inodemap, p1)
	f.inodemap[p2] = ino

	return 0
}

var _ fs.NodeUnlinker = (*ITNode)(nil)

func (f *ITNode) Unlink(ctx context.Context, name string) syscall.Errno {
	delete(f.inodemap, f.path+"/"+name)
	err := f.devfs.Remove(f.path + "/" + name)
	if err != nil {
		f.log.Error("FUSE Unlink failed", slog.String("file", f.path+"/"+name), slog.Any("error", err))
		return syscallErr(err)
	}

	f.log.Debug("FUSE Unlink success", slog.String("file", f.path+"/"+name))
	return 0
}

var _ fs.NodeRmdirer = (*ITNode)(nil)

func (f *ITNode) Rmdir(ctx context.Context, name string) syscall.Errno {
	return f.Unlink(ctx, name)
}
