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
		size:    6644,
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
