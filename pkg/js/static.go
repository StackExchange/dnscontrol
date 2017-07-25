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
		local:   "pkg/js/helpers.js",
		size:    11743,
		modtime: 0,
		compressed: `
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
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
