// Copyright (c) 2017 Gorillalabs. All rights reserved.

package middleware

// utf8 implements a primitive middleware that encodes all outputs
// as base64 to prevent encoding issues between remote PowerShell
// shells and the receiver. Just setting $OutputEncoding does not
// work reliably enough, sadly.
// type utf8 struct {
// 	upstream Middleware
// 	wrapper  string
// }

// func NewUTF8(upstream Middleware) (Middleware, error) {
// 	wrapper := "goUTF8" + utils.CreateRandomString(8)

// 	_, _, err := upstream.Execute(fmt.Sprintf(`function %s { process { if ($_) { [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($_)) } else { '' } } }`, wrapper))

// 	return &utf8{upstream, wrapper}, err
// }

// func (u *utf8) Execute(cmd string) (string, string, error) {
// 	// Out-String to concat all lines into a single line,
// 	// Write-Host to prevent line breaks at the "window width"
// 	cmd = fmt.Sprintf(`%s | Out-String | %s | Write-Host`, cmd, u.wrapper)

// 	stdout, stderr, err := u.upstream.Execute(cmd)
// 	if err != nil {
// 		return stdout, stderr, err
// 	}

// 	decoded, err := base64.StdEncoding.DecodeString(stdout)
// 	if err != nil {
// 		return stdout, stderr, err
// 	}

// 	return string(decoded), stderr, nil
// }

// func (u *utf8) Exit() {
// 	u.upstream.Exit()
// }
