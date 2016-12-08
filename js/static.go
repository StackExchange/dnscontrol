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
		size:    7179,
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
