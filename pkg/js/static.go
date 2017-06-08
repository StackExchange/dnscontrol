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
		size:    10861,
		modtime: 0,
		compressed: `
H4sIAAAAAAAA/+w6e3PbuPH/+1PscX6/iIwYynYSX4c+tlVt+cZTy/bISuobVdXAJCQh4WsAUI6bkz97
Bw+SICU5dqfpzN3Uf8gksNgXFvvA0ioYBsYpCbl1vLe3QhTCLJ1DAF/3AAAoXhDGKaLMh8nUlWNRymY5
zVYkwo3hLEEklQN7a40rwnNUxLxPFwwCmEyP9/bmRRpykqVAUsIJisk/se0oYg3Ku6g/wUGbC/G+PlbM
bTCyNli5xPejkpSdogS7/CHHboI5cjQ7ZA62GHQq9sQbBAFYw/7lh/6FpQit5a+QneKFEEag80EilUt8
+euCQO7LX82ikN6rJfbygi1tihfOsd4JXtBUItpg/jRl11oddk1J0TAEAFuKkM3lBARBAJ3s7hMOeceB
V6/A7pB8FmbpClNGspR1gKQKh2NsihjwmoAQwDyjCeIzzu0t805LNRHLX66axqYr7UQs/5Z2Unx/Kk1C
KabSr1MZuFzY4KUC8utHzdXXtZgOMxoxfzJ1hSVe14YoZrWljccXPuy7EiPDVGjCn0zXTeZymoWYsVNE
F8xOXG28prJ7PaFZwChcQpJFZE4wdcVeEg6EAfI8rwGrMfsQojgWQPeELzVeExBRih78kgEhUkEZWeH4
wYRSxiG2gi6wJJnyTCoiQhxVkOJszDzCzjR1O2kYTGk3thbvuJpZA44Zrtb3BVNbFgsN2MJuPkmD3MTd
1OPk07RSZQNwvYvwlZRzC+WZh79wnEaadU+I7iabEpir+JJm92D9rT+6PL/82decVLun/EaRsiLPM8px
5IPVhfJcQhcsUAYrxzVdZde1HOu9vV4PTts27cMJxYhjQHB6eaPxePCBYeBLDDmiKMEcUwaIlWYMKI0E
c8yr7XIDsRZQnl0lTrD7ZClGq00jEMD+MZCfTCfsxThd8OUxkG7XqbTX2EcDekKmrrGh600Ch4IAoosi
wSlvYjc2R0AnEEAFOCHTWq07TmPtu5QbUgFGOyANovdjcNb/cDG+Ae2mGCBgmEM2L0WvKQPPAOV5/CAf
4hjmBS8oLuOXJ/ANxKmXB5lnNfJ7EscQxhhRQOkD5BSvSFYwWKG4wEwQNHdSrypD7GYc3L5X31SluZdS
FaZOnTIWKr2Mxxf2yvHhBnNph+PxhSSprFTZocGzAjfirjiiN5ySdGGvHMfYTghk7pIuxtlpQZH0PSvH
DMTavZe4bWrKQD3OYwhgZbBbcbEFcX0IEsTDJRYqXHny2e79w/571HXsCUuW0X36MP2T8389zYqQoVoR
QFrEsSGF8hcrefIJgzTjgMRmkggiTVszYxmCFSnhEIDFrDaJyeHUwK7h6jkzFEMgfALD5ymvVh9MnUrM
QkRpi1n+gQtWYvlH+y5YS8t/e7S/r9mYWJE1hQAKbwmv4fBdOXqvRyN4DT+Wg6kx+Ha/HH0wR4/ea9Ze
B1BMBPfTRoRflWetCrMN0yrPWWlicky5QeNQmGu/j51FjbPi1UlBy9yULEb6ZpUpziVKsOXCvgMCJGUn
WZFKV7IPCUYpgyhLOxxE/p5Rnadg5RKMnMMzFwvTKtFrJGI5imNTORu5pF7ulIoqk8gSrcwjizTCc5Li
qGMoroKANwcv0ZaRVE0ED8I+NK6mZ+krFkleZmVDHWWZ53lOLZSGA5KboUxEPQhggXm1rHZj7qHzbV5R
FI0kXTtyrb7lltwIzE6T037/2cxWoN+Z337/aZYvzvs3uhxCdIH5t/iu4UEt+J7MC2Kae81dSwIhwsll
fzh4gQgG/PcXQRJ7UoReD25GHxU/OSUZJfzBvcdkseSuSB6fJ1SFAiocoJGAxNISNVzi8LMI7Pak9ogu
iOfLIrkTBchTzwp+6ta5jgvWzegj4C85DjmD5zFjOc/U+/uX6F3oIlL8WC48hxEXNjdlfDt+gVFV0N/f
pMa3428Z1PC2ZU/PkqFcZSjr3zKarcYxvN1tGy+1hrfPUJmsdmTSWtIxKjpTn4K1ykx2mEOloloD8olJ
GZkLEWahUycZqK5+4Ce1qHxvJ4W2XGrEzS01VQNBq5yS9H5QEBMylaRFdu40i9yaVteCN9XOgNUl3Sol
DTNKcchloWo5Rilq2tblS8LF5X8tVlw+HSgE4/3h4GYw+jgYmQKYzLYAWkx/I6ExEzJpd82rL4nK1//X
22yrvl3jFKVMvM44uov1daRwSYL+ZBJn9z4cuLAki6UPh66o0v+CGPbh7dQFNf2unH4vp8+vfTiaThUa
ecFjHcAjHMIjvIXHY3gHj/AeHgEe4UgUG2KDYpJiVUDumVYZCJuEn6DF5LYaUsLnELRhq4pcAEjuIACS
e/KxLqfka8PSjRskNdmy8hLXzEtQrkDcar+I87W8QSySwyjjNnHWjvcpI6ltuaa945jh7YjLlYr68cYR
MYQSO1KJJV4agomBJ0ST05vCaZyVeOL9PyagRm6IKLnYLaQoaQOY6PmKZu7F2b3jbg4Lg6zHNfd7hoLl
sypepfHp6/HsXssAj2A5QgzBgxZVAer5Y7DKe5rz4fXVaDwbj/qXN2dXo6E6VLGs65QV1pc/1RF8/iKX
8/hZjkF1CUIIWkGnTcpywfqzVaGv1Kr+vnZaR6jjt/2FyaWznjqNACG4bW44xaG+GeE83txjpcTrD6Of
B7ahIDWgBYy8v2Kcf0g/p9l9CgHMUcxw6WyvZhuLq7Ed6zktcMMjtmMDcxlHdFsU2XrJJYGP5T3Xziuu
Ok0oA+dmCStgmnf65lbKdsZG5NEkhLeda6cvo6xOkxBjRYKFc0RRRDFjHqhWCgfCvcZlhcqsbB2LTN41
2vrIapjNJpUwv69m92V3aHKFPfjmbUadqclmh26R6K7N9t5FhEMSYbhDDEeQparxU8K/gbNWB4OpDgZf
Yp1NAGLyrcwH6qVXW7sVArbRsZCwSnM+nJ/B8LbGrDQvt6MUrL5eM/Zuw55UMiYtZoc1gXH/LOAmZNqY
e14TBRKb4tBwvPCCbgYo8UtrqtyGvIxWF1xsc4GU3auA4dUrMJo19UQ7JlUcG2sbfUJj6ebC9cZQ1YsR
7mmjEfN8qJa29BlKZAe07uneWlu0J3CWdiG2cSviTS2EWcoykQZlC7vuCw13NoQst+oHuWDZN59JnpN0
8YNjtUXZGn8jT7d2yhZy2GySUhzucFmqOq691lN3DuZxeIYnqf1Eu9z2G2W33yi+n/I8vxVnc3Z+OxzY
PCaJ48MZCrm8liYMwizCkBVcnD7CGYhIV+6J9z+38/t0O78Z79DrkRzqbzgqy2Qwp1kCS85zv9djHIWf
sxWm8zi798Is6aHeHw723//4br93cHhwdLQvMr4VQeWCT2iFWEhJzj10lxVcronJHUX0oXcXk1ybibfk
iZF8X9tRxp09ow0NAUQZ91geE253vE5TClv+daPJ/tR5ffj+yOmKl4OpY7wdNt7eTp3WlyNlsVMkJWEy
F2+y4VH1OxzzcyVJ22p8CtRqLAlsm0vSImklZpHK3f7/8P3RlvT1raiz/yiP/5s3yoyNrotgEYaIL715
nGVU0OwJOWvzMLBDFzpeB7oQbenQRMfVTXqcFdE8RhQDiglimPnqOhFz2eTm4hRLJkkakRWJChSXnxh4
8luwk7PZ9ejq9pfZ1dmZiBSdsEI5y2n25aHjQyebzzvrY8mjqDHEMESEicIlaqO53I0lLZEYaHC6DcvZ
h4uLnXjmRRwrTCWW7giReFGkNTYxg+mb8isPUx3+Xi2D7ktm87mKUyknVbcfbKN16fhNBnUHf6fWZnpd
rb0tVNNNorvIbNdqg4rQrjKKDzfjq6EL16Orj+engxHcXA9Ozs/OT2A0OLkancL4l+vBjdFeOZuNBqfn
o8HJ2GY0dCFiz7tCE4eI0dAjaYS/XM3llQX8EATw5gB+/VWg2Ta19Z7Tojgi8iqT0VB+/BIxDknBVH90
iVYYwixJENu45oSNDk4tj+WKEp3RsGu5VlfIVVXLpvjjwfD6d6eDhlBPKOJfAQAA///rGpJDbSoAAA==
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
