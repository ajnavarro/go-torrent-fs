package torrent

import (
	"io"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/anacrolix/torrent"
)

var _ fs.File = &torrentFile{}

type torrentFile struct {
	*io.SectionReader
	io.Closer
	fi fs.FileInfo
}

func (f *torrentFile) Stat() (fs.FileInfo, error) {
	return f.fi, nil
}

var _ fs.ReadDirFile = &dir{}

type dir struct {
	files  []*fileAndPath
	name   string
	offset int
}

func (d *dir) Stat() (fs.FileInfo, error) {
	return newDirFileInfo(d.name), nil
}

func (d *dir) ReadDir(count int) ([]fs.DirEntry, error) {
	n := len(d.files) - d.offset
	if count > 0 && n > count {
		n = count
	}
	if n == 0 {
		if count <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}

	list := make([]fs.DirEntry, n)
	for i := range list {
		fp := d.files[d.offset+i]

		if fp.file == nil {
			list[i] = newDirEntry(newDirFileInfo(fp.path))
		} else {
			list[i] = newDirEntry(newTorrentFileInfo(fp.file))
		}
	}
	d.offset += n
	return list, nil
}

func (d *dir) Read([]byte) (int, error) { return 0, fs.ErrInvalid }
func (d *dir) Close() error             { return nil }

var _ fs.DirEntry = &dirEntry{}

type dirEntry struct {
	fi fs.FileInfo
}

func newDirEntry(fi fs.FileInfo) *dirEntry {
	return &dirEntry{
		fi: fi,
	}
}

func (de *dirEntry) Name() string {
	return de.fi.Name()
}

func (de *dirEntry) IsDir() bool {
	return de.fi.IsDir()
}

func (de *dirEntry) Type() fs.FileMode {
	return de.fi.Mode().Type()
}

func (de *dirEntry) Info() (fs.FileInfo, error) {
	return de.fi, nil
}

var _ fs.FileInfo = &dirFileInfo{}

func newDirFileInfo(name string) *dirFileInfo {
	return &dirFileInfo{name: name}
}

type dirFileInfo struct {
	name string
}

func (fi *dirFileInfo) Name() string {
	return path.Base(fi.name)
}
func (fi *dirFileInfo) Size() int64 {
	return 0
}
func (fi *dirFileInfo) Mode() fs.FileMode {
	return fs.ModeDir | 0555
}
func (fi *dirFileInfo) ModTime() time.Time {
	return time.Time{}
}
func (fi *dirFileInfo) IsDir() bool {
	return true
}
func (fi *dirFileInfo) Sys() interface{} {
	return nil
}

var _ fs.FileInfo = &torrentFileInfo{}

func newTorrentFileInfo(f *torrent.File) *torrentFileInfo {
	return &torrentFileInfo{
		f: f,
	}
}

type torrentFileInfo struct {
	f *torrent.File
}

func (fi *torrentFileInfo) Name() string {
	return path.Base(fi.f.Path())
}
func (fi *torrentFileInfo) Size() int64 {
	return fi.f.Length()
}
func (fi *torrentFileInfo) Mode() fs.FileMode {
	return os.ModeTemporary
}
func (fi *torrentFileInfo) ModTime() time.Time {
	return time.Time{}
}
func (fi *torrentFileInfo) IsDir() bool {
	return false
}
func (fi *torrentFileInfo) Sys() interface{} {
	return nil
}
