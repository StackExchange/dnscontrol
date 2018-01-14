package JsonConfigReader

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

var tests = map[string]string{
	`{
		// a
		"x": "y", // b
		"x": "y", // c
	}`: `{
		    
		"x": "y",     
		"x": "y"      
	}`,
	`{
		/*
		multiline comment
		*/
		"x": "y", // b
		"x": "y", // c
	}`: `{
		                           
		"x": "y",     
		"x": "y"      
	}`,
	`{
		/*
		multiline comment with special chars in comment * * /* \* / \\ end
		*/
		"x": "y", // b
		"x": "y", // c
	}`: `{
		                                                                            
		"x": "y",     
		"x": "y"      
	}`,

	`// serve a directory
	"l/test": [
		{
		"handler": "fs",
		"dir": "../",
		// "strip_prefix": "",
		},
	],`: `                    
	"l/test": [
		{
		"handler": "fs",
		"dir": "../" 
		                      
		} 
	],`,

	`[1, 2, 3]`:                   `[1, 2, 3]`,
	`[1, 2, 3, 4,]`:               `[1, 2, 3, 4 ]`,
	`{"x":1}//[1, 2, 3, 4,]`:      `{"x":1}               `,
	`//////`:                      `      `,
	`{}/ /..`:                     `{}/ /..`,
	`{,}/ /..`:                    `{ }/ /..`,
	`{,}//..`:                     `{ }    `,
	`{[],}`:                       `{[] }`,
	`{[,}`:                        `{[ }`,
	`[[",",],]`:                   `[["," ] ]`,
	`[",\"",]`:                    `[",\"" ]`,
	`[",\"\\\",]`:                 `[",\"\\\",]`,
	`[",//"]`:                     `[",//"]`,
	`[]/* missing close at end`:   `[]                       `,
	`[]/* missing close at end *`: `[]                         `,
	`[]/* 
	missing close at end`: `[]                         `,
	`[",//\"
		"],`: `[",//\"
		"],`,
}

func TestMain(t *testing.T) {
	for a, b := range tests {
		buf := &bytes.Buffer{}
		io.Copy(buf, New(strings.NewReader(a)))
		a = buf.String()
		if a != b {
			a = strings.Replace(a, " ", ".", -1)
			b = strings.Replace(b, " ", ".", -1)
			t.Errorf("reader failed to clean json: \nexpected: `%s`, \n      got `%s`", b, a)
		}
	}
}
