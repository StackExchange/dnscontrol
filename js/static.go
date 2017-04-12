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
		size:    7563,
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
