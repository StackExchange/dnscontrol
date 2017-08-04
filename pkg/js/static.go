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
		size:    14553,
		modtime: 0,
		compressed: `
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
Xf5vAAAA//8TlsMQ2TgAAA==
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
