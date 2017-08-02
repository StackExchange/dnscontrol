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
		size:    14687,
		modtime: 0,
		compressed: `
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
tFrT6v8GAAD//3lQzAtfOQAA
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
