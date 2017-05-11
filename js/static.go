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
		size:    8151,
		modtime: 0,
		compressed: `
H4sIAAAAAAAA/7xZbW/juPF/708xJ+C/lv5R5CR7mxbyuaibxIdFEydInDaFYRiMRNvclUSBpJJLF85n
L/ggiZLsTRbodl9kLXE485sHzgxHTsExcMFIJJxhr/eEGEQ0W8EIvvUAABheEy4YYjyE+cJX7+KML3NG
n0iMG69pikimXvS2hleMV6hIxJitOYxgvhj2eqsiiwShGZCMCIIS8m/selpYQ/I+6d9B0EYhn7dDDa4D
ZGtBmeLn21KUm6EU++Ilx36KBfIMHLICV770KnjyCUYjcK7G0/vxpaMFbdVfqTvDa6mMZBeCYqq2hOqv
D5J5qP4aiFL7oNY4yAu+cRlee0PjCVGwTDHqgD/P+I0xh1tL0jIsBcBVKtCVWoDRaAR9+vgFR6LvwYcP
4PZJvoxo9oQZJzTjfSCZ5uFZTpEvgiYhjGBFWYrEUgh3x7rXMk3M8x83TcPp2joxz9+yToafz1VIaMNU
9vWqAFcbG1gqorD+aVB928rliLKYh/OFLyPxpg5EuWoibTa7DOHIVxw5ZtIS4XyxbYLLGY0w5+eIrbmb
+iZ4bWMPBtKygFG0gZTGZEUw86UviQDCAQVB0KA1nEOIUJJIomciNoavTYgYQy9hCUCqVDBOnnDyYlPp
4JCuYGusRGaCKkPESKCKUp6NZUD4xEh300bAlHHjGvWG1coWcMJxtX8sQe3YLC3gyrj5ogKyy7tpx/mX
RWXKBuF2n+BrpecOycsA/yFwFhvogVTdT7sa2LvEhtFncP45vp1+nv4eGiSV93TeKDJe5DllAschOAdQ
nks4AAd0wKr3Rq6O61qPba83GMB5O6ZDOGMYCQwIzqd3hk8A9xyD2GDIEUMpFphxQLwMY0BZLMHxoI7L
DmOjoDq7Wp3R/pOlgVZOIzCCoyGQ3+wkHCQ4W4vNEMjBgVdZr+FHi3pOFr7l0G1XwIkUgNi6SHEmmtwt
50jqFEZQEc7JojbrntNY5y6dhnSBMQnIkBh/XEzG95ezOzBpigMCjgXQVal6LRkEBZTnyYv6kSSwKkTB
cFm/AsnvQp56dZAFrZk/kySBKMGIAcpeIGf4idCCwxNKCsylQNuTZldZYrt1cLev3jSl7UtlCtumXlkL
tV1ms0v3yQvhDgsVh7PZpRKpo1THoYVZkzfzc7noMhsEC4RIYARPTXnnVQpuiC19UIpX7/QRsQxm792D
IW4YIqgzfguKBmPVZqesX1OUYseHIw8kScbPaJGpODmCFKOMQ0yzvgDZnFFmihDW/rYKSmBvzqgo444Z
JnI7ShJbu06jYLZ7ZZNQdgglW9UkFFmMVyTDcb8+qzUFHB7bvc9b1rIq5lxiWMhconk13TjWEEleltwr
k0J5EARerZShA5LbeUqmNBjBGotqWx2j/on3NlYUx7dKrhv7ztjxSzSSs9dEOh6/G2xF+pPxjsffh3z5
eXxnel3E1li8hbumB73hZ4KXwgx6g66lgVThbDq+uvgBFSz6n6+CEvZdFWRifJj9AP6K+uejnz3M3sJ+
9aDB5IxQRsTL+3Qod0G1raVMtMHRV1lV3LnszO4EI9naB/l7WqSPsvut3y/8uqD64Fw9AP4jx5HgsE+K
473TZB/fYTLVNaniV8qxOkPbnhKa44PtPB9aJq1MVFtA/eJKRy4vFjzy6ssoqrso+E1vKp+tJK2aUVdt
tVL0jt6swaDVlil5v2iKOVko0bLKe81muZZ14MBh5RlwDsiBI28rskRFlDEcCdXwOp7V0tqxNf2RzDT9
n6Wl6fdzkgQ+vrq4u7j9x8WtrYANtkXQAv1G7bRrv4q75hVasQrN/9tdsVXf0gVDGZePS4EeEzPWkClJ
yp/PE/ocwrEPG7LehHDiy27/b4jjED4ufNDLv5bLn9Ty55sQThcLzUZdFJ1jeIUTeIWP8DqEX+EVPsEr
wCucOj3toIRkWDeiPTsqRzIm4TdogdzViyr6HEZt2qqzlwQKHYyA5IH6OaxOkXpsRLp1E9WLrSgveS2D
FOWaxK/8Rbxv5SSiSE9iKlzibb3gCyWZ6/h2vMtr427G5U4tfdg5IpZS0iOVWvKhoZh88R3V1HJXOcOz
Uk8+/9cUNMwtFRWK/UrKq/QI5ma9kpkHCX32/O5rGZD1e4O+ZxlY/dajQRV8ZsxGn40O8AqOJ9WQGIyq
mtCsD8Ep73ufr26ub2fL2e14eje5vr3ShypB0lI6CutLZHUE37/JFyJ5V2LQ08ZIXmwbRactyvHB+atT
sa/Mqv9967eOUD9s5wsbpbddeI0CIdE2Hc5wZC5oQiRdH2sj3tzf/n7hWgbSL4yCcfB3jPP77GtGnzMY
wQolHJfJ9nrZ2Vy927NfsAI3MmK7NnCfC8R2VZGdl2VFPFT35b1X5bpNKAtn97YkaZqzQduVaizaqTxG
hMy2K5P0VZU1bRLivEixTI4ojhnmPAA9khVARFAlirqzck0tsrEbtvWRNTTdYbcMv2/2FHd/afJlPIT2
xbnu1NTQ1IxazfR39ww0xhGJMTwijmOgmR4gl/SHMGlNQrmehMo7v+4mAHH1VPYD9dbrnVNPSduYfCpa
bbkQPk/g6qHmrC2v3FEqVhnc9l0nnnQzpiJmTzSBNceSdHOyaKy9bxgLqctwZCVe+IGpKGj1y2iq0oYa
anHVmfPuBqV7UBHDhw9gDX3rhXZNqhBbexvfG6yt3Y3bzqtqpivTU2eg+36qlrXMGUrVl5T629CDs8N6
kmcZF9KNOxl3rRDRjFPZBtG1W8+Xr/YOlh2/miv74Lh3X0mek2z9i+e0VdlZf+PAjIjLT1FR82MLw9FQ
p2KSQ/21pypSHFaMprARIg8HAy5Q9JU+YbZK6HMQ0XSABn8+Pvr0p1+PBscnx6enRzKnPxFUbviCnhCP
GMlFgB5pIdSehDwyxF4GjwnJTfwFG5Fa5fXGjanwetbAGkYQUxHwPCHC7Qf9phau+ncQz48W3v+ffDr1
DuTD8cKznk4aTx8XXusbU9nOFGkpmKzkk5qeVcMzz/6wqWQ7jY+GZSTpu63i1t2SFWkr9cY6O//fyafT
HQXqo+yk/6LyyuGhPh/WCE9ChCskNsEqoZRJmQOpZx0eFnc4gH7QhwOId4z7YhMKcJbQIl4liGFACUEc
81B9Cj6bLG9urx/+tbyeTGSp6EcV4TJn9I+Xfgh9ulr1t0MluJ7jDuhqFbR4TPezyEoOiscdEZgDRyss
k5OiavGa3F9e7uW2KpJE8+vwYogk6yLT3O4uZoadmbgvZ7f3FzCyIQ+r+xXN4GyiwejSlgnC6mH0boaT
8eVdk+Nkolg+Ik5kkXsBBBk9pHnQ+08AAAD//5rwOkTXHwAA
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
