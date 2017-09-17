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
		size:    14682,
		modtime: 0,
		compressed: `
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
AAD//4nEKeNaOQAA
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
