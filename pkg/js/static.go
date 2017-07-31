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
		size:    14686,
		modtime: 0,
		compressed: `
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
/odWa1r93wAAAP//0yHyLV45AAA=
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
