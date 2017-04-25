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

// _escFS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func _escFS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// _escDir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func _escDir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// _escFSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func _escFSByte(useLocal bool, name string) ([]byte, error) {
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

// _escFSMustByte is the same as _escFSByte, but panics if name is not present.
func _escFSMustByte(useLocal bool, name string) []byte {
	b, err := _escFSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// _escFSString is the string version of _escFSByte.
func _escFSString(useLocal bool, name string) (string, error) {
	b, err := _escFSByte(useLocal, name)
	return string(b), err
}

// _escFSMustString is the string version of _escFSMustByte.
func _escFSMustString(useLocal bool, name string) string {
	return string(_escFSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/helpers.js": {
		local:   "js/helpers.js",
		size:    7946,
		modtime: 0,
		compressed: `
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
HwAA
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
