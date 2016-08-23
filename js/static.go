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

// FS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func FS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// Dir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func Dir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// FSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func FSByte(useLocal bool, name string) ([]byte, error) {
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

// FSMustByte is the same as FSByte, but panics if name is not present.
func FSMustByte(useLocal bool, name string) []byte {
	b, err := FSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// FSString is the string version of FSByte.
func FSString(useLocal bool, name string) (string, error) {
	b, err := FSByte(useLocal, name)
	return string(b), err
}

// FSMustString is the string version of FSMustByte.
func FSMustString(useLocal bool, name string) string {
	return string(FSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/helpers.js": {
		local:   "js/helpers.js",
		size:    6427,
		modtime: 0,
		compressed: `
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
`,
	},

	"/tester.html": {
		local:   "js/tester.html",
		size:    953,
		modtime: 0,
		compressed: `
H4sIAAAJbogA/2yST2/bPAzG7/4UhN73YCOr1TrrOjSWDwM6YD10w1AM6IocHFmJ6SliJsnZsiLffZD/
NAmykynqx4cPKee1X+siymtVVkUEAJA7aXHjwVkpWO39xt1yLqlSafOzVXaXSlrzPrzI0iydpms0aeMY
oPFqZdHvBHN1mV2/uyiz6erqwb+t/N3T5+Zm8XWyfWxvJo/vv325mtbt+vvHD/cPn57o7p4EA2nJObK4
QiNYacjs1tQ6VuS8t1REOe995guqdoPdelpUxkky3pKGxoFXziub83o6EF799qVVJWAlGBoGln45wa4v
GUjSTrDs8jJ0GblR2PIhWrTekwEyUqP8IVjocO/ihBWPyvmc9/dnZRVuu5ZWuVb7bpAKt0UUna34lvPW
VMo6SVY1LiW7OkpcDAs+XsSJgtIbZd2/kSJatkZ6JAOj7ZfOHhr0WGr8o+Jk1mW2pYXw0CDg/5j9h4Yl
6bbU47Uk40irVNMqDtiQVgE5Op9iZjmkg+K4iCQNf13cODJZF3VcMov2gTwYPgAhSuDl1Sa+Aas8CGCs
lw+niQCWt7oYUkuyMSBgrzNWn8AaCzbBCbuFoaZbzDIGv9soWnaFzzgHIQQwWjRKepaM5afunnGeHDSU
duqYe8b57Lw9D/37/P50Cn4YwyrfWhM+s2gfvb7v3wAAAP//jV4zrLkDAAA=
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
