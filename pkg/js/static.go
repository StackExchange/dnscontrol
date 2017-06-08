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
		size:    8979,
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
