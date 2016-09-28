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
		size:    6729,
		modtime: 0,
		compressed: `
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
AAA=
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
