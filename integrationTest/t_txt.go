package main

import "github.com/StackExchange/dnscontrol/v2/providers"

func init() {
	tests = append(tests, []*TestCase{

		// TXT (single)
		reset(),
		tc("Create a TXT", txt("foo", "simple")),
		tc("Change a TXT", txt("foo", "changed")),
		reset(),
		tc("Create a TXT with spaces", txt("foo", "with spaces")),
		tc("Change a TXT with spaces", txt("foo", "with whitespace")),
		tc("Create 1 TXT as array", txtmulti("foo", []string{"simple"})),
		reset(),
		tc("Create a 255-byte TXT", txt("foo", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")),

		// FUTURE(tal): https://github.com/StackExchange/dnscontrol/issues/598
		// We decided that handling an empty TXT string is not a
		// requirement. In the future we might make it a "capability" to
		// indicate which vendors fully support RFC 1035, which requires
		// that a TXT string can be empty.
		//
		// TXT (empty)
		reset(not("DNSIMPLE")),
		tc("TXT with empty str", txt("foo1", "")),

		// TXTMulti
		reset(requires(providers.CanUseTXTMulti)),
		tc("Create TXTMulti 1",
			txtmulti("foo1", []string{"simple"}),
		),
		tc("Create TXTMulti 2",
			txtmulti("foo1", []string{"simple"}),
			txtmulti("foo2", []string{"one", "two"}),
		),
		tc("Create TXTMulti 3",
			txtmulti("foo1", []string{"simple"}),
			txtmulti("foo2", []string{"one", "two"}),
			txtmulti("foo3", []string{"eh", "bee", "cee"}),
		),
		tc("Create TXTMulti with quotes",
			txtmulti("foo1", []string{"simple"}),
			txtmulti("foo2", []string{"o\"ne", "tw\"o"}),
			txtmulti("foo3", []string{"eh", "bee", "cee"}),
		),
		tc("Change TXTMulti",
			txtmulti("foo1", []string{"dimple"}),
			txtmulti("foo2", []string{"fun", "two"}),
			txtmulti("foo3", []string{"eh", "bzz", "cee"}),
		),
		reset(requires(providers.CanUseTXTMulti)),
		tc("3x255-byte TXTMulti",
			txtmulti("foo3", []string{"XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", "YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY", "ZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ"}),
		),

		// Close out the previous test.
		reset(),
	}...,
	)
}
