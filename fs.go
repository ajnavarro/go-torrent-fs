package torrent

import (
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/anacrolix/torrent"
)

var _ fs.FS = &Torrent{}

type Torrent struct {
	t *torrent.Torrent

	loadOnce sync.Once
	fileList *fileList
}

func New(t *torrent.Torrent) *Torrent {
	return &Torrent{
		t:        t,
		fileList: newFileList(),
	}
}

func (vfs *Torrent) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	vfs.load()

	// special case for root
	if name == "." {
		files := vfs.fileList.Dir(name)
		return &dir{files: files, name: "."}, nil
	}

	f, ok := vfs.fileList.Get(name)

	isDir := f == nil && ok

	if isDir {
		files := vfs.fileList.Dir(name)
		return &dir{files: files, name: name}, nil
	}

	if f == nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	// TODO workaround to avoid this error: ReadAll vs fs.ReadFile: different data returned
	// It might be a bug from torrent file reader implementation.
	tr := newReadAtWrapper(f.NewReader())
	sr := io.NewSectionReader(tr, 0, f.Length())

	return &torrentFile{
		SectionReader: sr,
		Closer:        tr,
		fi:            newTorrentFileInfo(f),
	}, nil
}

func (vfs *Torrent) load() {
	vfs.loadOnce.Do(func() {
		<-vfs.t.GotInfo()
		dirs := make(map[string]bool)
		for _, file := range vfs.t.Files() {
			dir := file.Path()
			for {
				dir, _ = path.Split(dir)
				dir = strings.Trim(dir, "/")
				if dir == "" {
					break
				}

				dirs[dir] = true
			}

			vfs.fileList.Add(file.Path(), file)
		}

		for k := range dirs {
			vfs.fileList.Add(k, nil)
		}

		vfs.fileList.Sort()
	})
}
