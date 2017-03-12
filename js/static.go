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
		size:    7196,
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
