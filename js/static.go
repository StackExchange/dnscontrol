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
		size:    7000,
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
