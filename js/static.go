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
		size:    8296,
		modtime: 0,
		compressed: `
H4sIAAAAAAAA/9w5W2/bytHv+hVzCHwR+YmhL0ncgoqKqrZ8YNSSDUk+dSEIxppcSZvwht2lfNwc+bcX
eyG5JKXYAZo+NA+OdnfuMzszO7RyhoFxSgJu9TudLaIQpMkKBvCtAwBA8ZowThFlPiyWrtwLE/aQ0XRL
QlzbTmNEErnR2WlaIV6hPOJDumYwgMWy3+ms8iTgJE2AJIQTFJF/YdtRzGqcD3H/jgRNKcR611fCtQTZ
GaJM8NO0YGUnKMYuf86wG2OOHC0OWYEtNp1SPLGCwQCs8XByN7y2FKOd/Ct0p3gtlBHkfJBEJYov/7og
iPvyrxZRaO9VGntZzjY2xWunrz3Bc5pIQi3hLxJ2q81hV5wUD0MBsKUK6UoewGAwgG76+AUHvOvAu3dg
d0n2EKTJFlNG0oR1gSSKhmM4RWx4dUAYwCqlMeIPnNt7zp2GaUKW/bhpak5X1glZ9pp1Evx0IUNCGaa0
r1MGuESsyVIC+dVPLdW3nTgOUhoyf7F0RSTeVoEoTnWkzefXPhy7kiLDVFjCXyx3deEymgaYsQtE18yO
XR28prGPjoRlAaNgA3EakhXB1BW+JBwIA+R5Xg1WU/YhQFEkgJ4I32i6JiCiFD37hQBCpZwyssXRswml
gkO4gq6xZJnwVBoiRByVkOJuPHiEXWrudlwLmCJubK1evzzZAY4YLvGHQqg9yMICtoibLzIg27Trdlx8
WZamrAHuDjG+kXru4fzg4d85TkItuidUd+O2BiYW39D0Cax/DKeTq8mvvpak9J7KG3nC8ixLKcehD1YP
insJPbBABazc13xVXFd67DqdoyO4aMa0D+cUI44BwcVkpul4cMcw8A2GDFEUY44pA8SKMAaUhEI45lVx
2SKsFZR3V6kzOHyzlKCl0wgM4LgP5LOZhL0IJ2u+6QPp9ZzSejU/GtALsnQNh+7aDE4FA0TXeYwTXqdu
OEdAxzCAEnBBlpVZD9zGKnepNKQKjE5AGkT7Y3Q5vLuez0CnKQYIGOaQrgrVK87AU0BZFj3LH1EEq5zn
FBf1yxP0RuLWy4vM04r4E4kiCCKMKKDkGTKKtyTNGWxRlGMmGJqe1FhFiW3Xwf2+etWUpi+lKUybOkUt
VHaZz6/trePDDHMZh/P5tWSpolTFoSGzAq/n5+LQpqYQ1OM8ggFs6/wuyhRcY1v4oGAv99QVMQxm4h6Q
IawZwqsyfkMUJYxRm62ifk1QjC0Xjh0QIAk7T/NExskxxBglDMI06XIQzVlKdRHCyt9GQfFM5CTlRdxR
TUSgoygytWs1ChrdKZqEokMoyMomIU9CvCIJDrvVXa0g4P2J2fu8Zi2jYi6EDEuRSxStuhuHSkSSFSV3
rFMo8zzPqZTScEAyM0+JlAYDWGNeolUx6p46r8uKwnAq+dqhaw0tt5BGUHbqkg6Hbxa2BP3J8g6H3xf5
+mo4070uomvMX5O7ggeF8DOFF8y09Fq6hgZChfPJcDz6ARUM+J+vgmT2XRVEYryf/4D8JfTPl35+P39N
9vG9EiajJKWEP79NhwILSrSGMsEGB19FVbEXojObcUqStQvi9ySPH0X3W+0v3aqgumCN7wH/nuGAMzjE
xXLeaLIPbzCZ7Jpk8Sv4GJ2haU8hmuWC6TwXGiYtTVRZQP5iUkcmHhYscKrHKKq6KPiskIq1kaRlM2pL
VCNF7+nNagQabZnk94uCWJClZC2qvFNvlitePQvel54Bq0d6lnitiBIVpJTigMuG13KMltaMrcmPZKbJ
fy0tTb6fk4Tgw/FoNpr+NpqaCpjCNgAaQr9SO83aL+Ou/oSWpHz9/25fbFWvdE5RwsTygaPHSI81REoS
/BeLKH3y4cSFDVlvfDh1Rbf/N8SwDx+WLqjjj8XxJ3l8devD2XKpyMiHonUCL3AKL/ABXvrwEV7gE7wA
vMCZ1VEOikiCVSPaMaNyIGISPkNDyH29qITPYNCELTt7ASClgwGQzJM/++UtkstapBsvUXXYiPKC1oMX
o0yBuKW/iPOtmETk8WmYcps4O8f7kpLEtlwz3sWzcT/hAlNx77euiKGU8EiplljUFBMb31FNHreV0zRL
9cT6P6agJm6oKKU4rKR4Sg9goc9LnpkXpU+O294WAVnta+k7hoHlbzUalMGnx2zpk9YBXsByhBpCBq2q
AtTnfbCK997V+PZmOn+YT4eT2eXNdKwuVYSEpVQUVo/I8gq+HcnlPHpTYlDTxkA8bGtFp8nKcsH6q1WS
L82q/n3rNq5Q12/mC1NKZ7d0agVCSFt3OMWBfqBxHrV9rIx4ezf9dWQbBlIbWsHQ+zvG2V3yNUmfEhjA
CkUMF8n25qGFXO4dwOc0x7WM2KwNzGUc0X1VZO9jWQL35Xv54FO5ahOKwtl+LQmY+mzQdKUci7Yqj2Yh
su1KJ31ZZXWbhBjLYyySIwpDihnzQI1kORDulYmi6qxsXYtM2TXZ6spqmPawW4TfN3OKe7g0uSIefPPh
XHVqcmiqR616+rt/BhrigIQYHhHDIaSJGiAX8O/hsjEJZWoSKt78qpsAxOSq6Acq1Ju9U08BW5t8Slhl
OR+uLmF8X1FWlpfuKBQrDW76rhVPqhmTEXMgmsCYYwm4BVnWzt42jIXYpjgwEi/8wFQUlPpFNJVpQw61
mOzMWRtB6u6VwPDuHRhD3+qgWZNKiQ3c2vcGA7WNuGttlTNdkZ5aA923QzWspe9QLL+kVN+G7q091hM0
i7gQbtxLuG2FIE1YKtqgdG1X8+XxwcGy5ZZzZRcse/aVZBlJ1r84VlOVvfU39PSIuPgUFdQ/tlAc9FUq
JhlUX3vKIsVgRdMYNpxn/tER4yj4mm4xXUXpkxek8RE6+vPJ8ac/fTw+Ojk9OTs7Fjl9S1CB8AVtEQso
ybiHHtOcS5yIPFJEn48eI5Lp+PM2PDbK660dptzpGANrGECYco9lEeF21+vWtbDlv164OF46/3/66czp
icXJ0jFWp7XVh6XT+MZUtDN5XDAmK7GS07NyeOaYHzYlb6v20bCIJPW2ldTaKEkeN1JvqLLz/51+OttT
oD6ITvovMq+8f6/uhzHCEyLCGPGNt4rSlAqeR0LPKjwM6tCDrteFHoR7xn2hDgU4v5vNb8Yu3E5vfru6
GE1hdjs6v7q8Oofp6PxmegHzf96OZsZU5vJhOrq4mo7O5zajgQshe9tzSJiL0cAjSRDlIWay/4Q//hAE
6pt736kWxSGRT1FGA/kRJGQc4pypUeoGbTEEaRwj1nqmQmvYU+lguaLFYjToWa7VE7qU3Y6p8nw0vv2f
0LumyGHl/x0AAP//n5W68mggAAA=
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
