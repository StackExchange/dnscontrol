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
		size:    7195,
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
