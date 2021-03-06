package torrent

import (
	"io"
	"os"
	"testing"
	"testing/fstest"

	"github.com/anacrolix/torrent"
	"github.com/stretchr/testify/require"
)

const testMagnet = "magnet:?xt=urn:btih:a88fda5954e89178c372716a6a78b8180ed4dad3&dn=The+WIRED+CD+-+Rip.+Sample.+Mash.+Share&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fwired-cd.torrent"

func TestTorrentFS(t *testing.T) {
	require := require.New(t)

	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = os.TempDir()
	cfg.Debug = true

	client, err := torrent.NewClient(cfg)
	require.NoError(err)

	defer client.Close()

	to, err := client.AddMagnet(testMagnet)
	require.NoError(err)

	t.Run("testFS", func(t *testing.T) {
		torr := New(to)

		err = fstest.TestFS(torr,
			"The WIRED CD - Rip. Sample. Mash. Share/01 - Beastie Boys - Now Get Busy.mp3",
			"The WIRED CD - Rip. Sample. Mash. Share/README.md",
			"The WIRED CD - Rip. Sample. Mash. Share/poster.jpg",
		)

		require.NoError(err)
	})

	t.Run("testTorrentFile", func(t *testing.T) {
		torr := New(to)

		f, err := torr.Open("The WIRED CD - Rip. Sample. Mash. Share/01 - Beastie Boys - Now Get Busy.mp3")
		require.NoError(err)

		_, ok := f.(io.Seeker)
		if !ok {
			require.Fail("reader must implement Seeker")
		}

		_, ok = f.(io.ReaderAt)
		if !ok {
			require.Fail("reader must implement ReaderAt")
		}
	})
}
