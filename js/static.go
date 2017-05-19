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
		size:    8320,
		modtime: 0,
		compressed: `
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
/mw0vvmfs0FNqf2G+HcAAAD//xD4Q9mAIAAA
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
