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
		size:    7758,
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
