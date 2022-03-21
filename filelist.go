package torrent

import (
	"sort"

	"github.com/anacrolix/torrent"
)

func newFileList() *fileList {
	return &fileList{
		filesByPath: make(map[string]*torrent.File),
	}
}

type fileList struct {
	files       []*fileAndPath
	filesByPath map[string]*torrent.File
	sorted      bool
}

func (fl *fileList) Add(path string, file *torrent.File) {
	fl.files = append(fl.files, &fileAndPath{file: file, path: path})
	fl.filesByPath[path] = file
	fl.sorted = false
}

func (fl *fileList) Sort() {
	sort.Slice(fl.files, func(i, j int) bool { return fileEntryLess(fl.files[i].path, fl.files[j].path) })
	fl.sorted = true
}

func (fl *fileList) Get(path string) (*torrent.File, bool) {
	file, ok := fl.filesByPath[path]
	return file, ok
}

func (fl *fileList) Dir(dir string) []*fileAndPath {
	if !fl.sorted {
		fl.Sort()
	}

	fis := fl.files
	i := sort.Search(len(fis), func(i int) bool {
		idir, _ := split(fis[i].path)
		return idir >= dir
	})
	j := sort.Search(len(fis), func(j int) bool {
		jdir, _ := split(fis[j].path)
		return jdir > dir
	})

	return fis[i:j]
}

func fileEntryLess(x, y string) bool {
	xdir, xelem := split(x)
	ydir, yelem := split(y)
	return xdir < ydir || xdir == ydir && xelem < yelem
}

func split(name string) (dir, elem string) {
	i := len(name) - 1
	for i >= 0 && name[i] != '/' {
		i--
	}
	if i < 0 {
		return ".", name
	}
	return name[:i], name[i+1:]
}

type fileAndPath struct {
	file *torrent.File
	path string
}
