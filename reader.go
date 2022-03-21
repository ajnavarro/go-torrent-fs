package torrent

import (
	"io"
	"sync"
)

var _ io.ReaderAt = &readAtWrapper{}
var _ io.Closer = &readAtWrapper{}

type readAtWrapper struct {
	mu sync.Mutex
	io.ReadSeekCloser
}

func newReadAtWrapper(r io.ReadSeekCloser) *readAtWrapper {
	return &readAtWrapper{ReadSeekCloser: r}
}

func (rw *readAtWrapper) ReadAt(p []byte, off int64) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Get actual position
	pos, err := rw.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	defer rw.Seek(pos, io.SeekStart)

	_, err = rw.Seek(off, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return readAtLeast(rw, p, len(p))
}

func readAtLeast(r io.Reader, buf []byte, min int) (n int, err error) {
	if len(buf) < min {
		return 0, io.ErrShortBuffer
	}
	for n < min && err == nil {
		var nn int

		nn, err = r.Read(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}
