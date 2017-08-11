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
		size:    13440,
		modtime: 0,
		compressed: `
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
O7++e1uMJYj/l2WDLP83AAD//yOh85uANAAA
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
