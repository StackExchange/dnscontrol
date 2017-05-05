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
		size:    8044,
		modtime: 0,
		compressed: `
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
OhuNr/8gzjVS+9n/PwAA//+W/LCWbB8AAA==
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
