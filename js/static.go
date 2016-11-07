package js

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/helpers.js": {
		local:   "js/helpers.js",
		size:    6952,
		modtime: 0,
		compressed: `
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
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
