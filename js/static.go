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
		size:    7946,
		modtime: 0,
		compressed: `
H4sIAAAAAAAA/7xZfW/bvBH/35/iHgGrpUVVXtpmg1wP85r0QbHYDRxnC2AYASPRNltJFEjKfrLC+ewD
XyRRkt0kwLr+kVri8e53L7w7npyCY+CCkUg4g15vgxhENFvCEH70AAAYXhEuGGI8hPnCV+/ijN/njG5I
jBuvaYpIpl70doZXjJeoSMSIrTgMYb4Y9HrLIosEoRmQjAiCEvIf7HpaWEPyIek/QdBGIZ93Aw2uA2Rn
QZng7bQU5WYoxb54zLGfYoE8A4cswZUvvQqefILhEJzxaHI7unK0oJ36K3VneCWVkexCUEzVllD99UEy
D9VfA1FqH9QaB3nB1y7DK29gPCEKlilGHfAXGb825nBrSVqGpQC4SgW6VAswHA6hTx++4Uj0PXjzBtw+
ye8jmm0w44RmvA8k0zw8yynyRdAkhCEsKUuRuBfC3bPutUwT8/z1pmk4XVsn5vlz1snw9kKFhDZMZV+v
CnC1sYGlIgrrnwbVj51cjiiLeThf+DISr+tAlKsm0mazqxBOfMWRYyYtEc4Xuya4nNEIc36B2Iq7qW+C
1zb28bG0LGAUrSGlMVkSzHzpSyKAcEBBEDRoDecQIpQkkmhLxNrwtQkRY+gxLAFIlQrGyQYnjzaVDg7p
CrbCSmQmqDJEjASqKOXZuA8I/2yku2kjYMq4cY16g2plBzjhuNo/kqD2bJYWcGXcfFMB2eXdtOP826Iy
ZYNwd0jwV6XnHsn3Af5D4Cw20AOpup92NbB3iTWjW3D+PZpOvkx+Dw2Syns6bxQZL/KcMoHjEJwjKM8l
HIEDOmDVeyNXx3Wtx67XOz6Gi3ZMh/CJYSQwILiY3Bg+AdxyDGKNIUcMpVhgxgHxMowBZbEEx4M6LjuM
jYLq7Gp1hodPlgZaOY3AEE4GQD7aSThIcLYS6wGQoyOvsl7Djxb1nCx8y6G7roAzKQCxVZHiTDS5W86R
1CkMoSKck0Vt1gOnsc5dOg3pAmMSkCEx/rj8PLq9mt2ASVMcEHAsgC5L1WvJICigPE8e1Y8kgWUhCobL
+hVIfpfy1KuDLGjNfEuSBKIEIwYoe4Sc4Q2hBYcNSgrMpUDbk2ZXWWK7dXC/r541pe1LZQrbpl5ZC7Vd
ZrMrd+OFcIOFisPZ7EqJ1FGq49DCrMmb+blcdJkNggVCJDCETVPeRZWCG2JLH5Ti1Tt9RCyD2XsPYIgb
hgjqjN+CosFYtdkp69cEpdjx4cQDSZLxT7TIVJycQIpRxiGmWV+AbM4oM0UIa39bBSWwN2dUlHHHDBO5
HSWJrV2nUTDbvbJJKDuEkq1qEoosxkuS4bhfn9WaAt6e2r3Pc9ayKuZcYljIXKJ5Nd040hBJXpbcsUmh
PAgCr1bK0AHJ7TwlUxoMYYVFta2OUf/Mex4riuOpkuvGvjNy/BKN5Ow1kY5GLwZbkf5ivKPRzyFffRnd
mF4XsRUWz+Gu6UFv+JXgpTCD3qBraSBV+DQZjS9foYJF/+tVUMJ+qoJMjHezV+CvqH89+tnd7Dnst9Or
V2CvqH899tvp1XPYx3caTM4IZUQ8vkyHchdU21rKRGscfZcV0Z3LrvJGMJKtfJC/J0X6IDv3+v3Cr5sB
H5zxHeA/chwJDoekON4LTfbuBSZTHZ8q3KUcq6u17SmhOT7YzvOhZdLKRLUF1C+udOTyUsQjr75Io7oD
hI96U/lsFRjVSLtqq1Ve9vSVDQatllLJ+01TzMlCiZYditds9GtZRw68rTwDzhE5cuRNS5bXiDKGI6Ga
dcez2nE7tiavyaqT/1tKnfw8n0rgo/HlzeX0X5dTWwEbbIugBfqZum/3LSrumtd/xSo0/+/2xVY9YRAM
ZVw+3gv0kJiRjEynUv58ntBtCKc+rMlqHcKZL28q/0Ach/Bu4YNefl8uf1DLX65DOF8sNBt1yXVO4QnO
4AnewdMA3sMTfIAngCc4d3raQQnJsG6ie3ZUDmVMwkdogdzXRyv6HIZt2upWIgkUOhgCyQP1c1CdIvXY
iHTrFq0XW1Fe8roPUpRrEr/yF/F+lFOUIj2LqXCJt/OCb5RkruPb8S6vvPsZlzu19EHniFhKSY9UasmH
hmLyxU9UU8td5QzPSj35/D9T0DC3VFQoDivJ6FaGh1mvZOZBQree330tA7J+b9D3LAOr33qsqYLPjAjp
1ugAT+B4Ug2JwaiqCc36AJzyrvplfP11OrufTUeTm89fp2N9qBIkLaWjsL4AV0fw5Zt8IZIXJQY9KY3k
pbxRdNqiHB+cvzsV+8qs+t+PfusI9cN2vrBReruF1ygQEm3T4QxH5nIpRNL1sTbi9e3090vXMpB+YRSM
g39inN9m3zO6zWAIS5RwXCbbr/edzdW7A/sFK3AjI7ZrA/e5QGxfFdl70VfEA3XXP3jNr9uEsnB2b3qS
pjnXtF2pRrqdymNEyGy7NElfVVnTJiHOixTL5IjimGHOA9DjZAFEBFWiqDsr19QiG7thWx9ZQ9Md1Mvw
+2FPoA+XJl/GQ2hf+utOTQ18zZjYTK73z29jHJEYwwPiOAaa6eF3Sf8WPremuFxPccUam24CEFdPZT9Q
b/26d2IraRtTW0WrLRfCl88wvqs5a8srd5SKVQa3fdeJJ92MqYg5EE1gzeAk3ZwsGmsvGyRD6jIcWYkX
XjHRBa1+GU1V2lADOa46c97doHQPKmJ48wasgXW90K5JFWJrb+NbibW1u3HXeVXNo2V66gyjX07VspY5
Q6n6ClR/17pz9lhP8izjQrpxL+OuFSKacSrbILpy69n4+OBQ3PGrmbgPjnvzneQ5yVa/eU5blb31Nw7M
eLv8jBY1PxQxHA10KiY51F+qqiLFYcloCmsh8vD4mAsUfacbzJYJ3QYRTY/R8V9PTz785f3J8enZ6fn5
iczpG4LKDd/QBvGIkVwE6IEWQu1JyAND7PH4ISG5ib9gLVKrvF67MRVezxq2wxBiKgKeJ0S4/aDf1MJV
/47i+cnC+/PZh3PvSD6cLjzr6azx9G7htb6Ple1MkZaCyVI+qclfNfjz7I+ySrbT+OBZRpK+2ypu3S1Z
kbZSb6yz85/OPpzvKVDvZCf9N5VX3r7V58MaP0qIMEZiHSwTSpmUeSz1rMPD4g5H0A/6cATxnlFlLE3y
3wAAAP//pDaD0gofAAA=
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
