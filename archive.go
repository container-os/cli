package docker

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dotcloud/docker/utils"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type Archive io.Reader

type Compression uint32

const (
	Uncompressed Compression = iota
	Bzip2
	Gzip
	Xz
)

func DetectCompression(source []byte) Compression {
	for _, c := range source[:10] {
		utils.Debugf("%x", c)
	}

	sourceLen := len(source)
	for compression, m := range map[Compression][]byte{
		Bzip2: {0x42, 0x5A, 0x68},
		Gzip:  {0x1F, 0x8B, 0x08},
		Xz:    {0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00},
	} {
		fail := false
		if len(m) > sourceLen {
			utils.Debugf("Len too short")
			continue
		}
		i := 0
		for _, b := range m {
			if b != source[i] {
				fail = true
				break
			}
			i++
		}
		if !fail {
			return compression
		}
	}
	return Uncompressed
}

func (compression *Compression) Flag() string {
	switch *compression {
	case Bzip2:
		return "j"
	case Gzip:
		return "z"
	case Xz:
		return "J"
	}
	return ""
}

func (compression *Compression) Extension() string {
	switch *compression {
	case Uncompressed:
		return "tar"
	case Bzip2:
		return "tar.bz2"
	case Gzip:
		return "tar.gz"
	case Xz:
		return "tar.xz"
	}
	return ""
}

func Tar(path string, compression Compression) (io.Reader, error) {
	return CmdStream(exec.Command("tar", "-f", "-", "-C", path, "-c"+compression.Flag(), "."))
}

func Untar(archive io.Reader, path string) error {

	buf := make([]byte, 10)
	if _, err := archive.Read(buf); err != nil {
		return err
	}
	compression := DetectCompression(buf)
	archive = io.MultiReader(bytes.NewReader(buf), archive)

	utils.Debugf("Archive compression detected: %s", compression.Extension())

	cmd := exec.Command("tar", "-f", "-", "-C", path, "-x"+compression.Flag())
	cmd.Stdin = archive
	// Hardcode locale environment for predictable outcome regardless of host configuration.
	//   (see https://github.com/dotcloud/docker/issues/355)
	cmd.Env = []string{"LANG=en_US.utf-8", "LC_ALL=en_US.utf-8"}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, output)
	}
	return nil
}

// CmdStream executes a command, and returns its stdout as a stream.
// If the command fails to run or doesn't complete successfully, an error
// will be returned, including anything written on stderr.
func CmdStream(cmd *exec.Cmd) (io.Reader, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	pipeR, pipeW := io.Pipe()
	errChan := make(chan []byte)
	// Collect stderr, we will use it in case of an error
	go func() {
		errText, e := ioutil.ReadAll(stderr)
		if e != nil {
			errText = []byte("(...couldn't fetch stderr: " + e.Error() + ")")
		}
		errChan <- errText
	}()
	// Copy stdout to the returned pipe
	go func() {
		_, err := io.Copy(pipeW, stdout)
		if err != nil {
			pipeW.CloseWithError(err)
		}
		errText := <-errChan
		if err := cmd.Wait(); err != nil {
			pipeW.CloseWithError(errors.New(err.Error() + ": " + string(errText)))
		} else {
			pipeW.Close()
		}
	}()
	// Run the command and return the pipe
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return pipeR, nil
}

// NewTempArchive reads the content of src into a temporary file, and returns the contents
// of that file as an archive. The archive can only be read once - as soon as reading completes,
// the file will be deleted.
func NewTempArchive(src Archive, dir string) (*TempArchive, error) {
	f, err := ioutil.TempFile(dir, "")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(f, src); err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	return &TempArchive{f, size}, nil
}

type TempArchive struct {
	*os.File
	Size int64 // Pre-computed from Stat().Size() as a convenience
}

func (archive *TempArchive) Read(data []byte) (int, error) {
	n, err := archive.File.Read(data)
	if err != nil {
		os.Remove(archive.File.Name())
	}
	return n, err
}
