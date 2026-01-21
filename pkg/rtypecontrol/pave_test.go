package rtypecontrol

import "testing"

func TestPaveArgs(t *testing.T) {
	tests := []struct {
		name     string
		dataArgs []any
		dataRule string
		wantErr  bool
	}{
		// String tests
		{
			name:     "string",
			dataArgs: []any{"one"},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "int to string",
			dataArgs: []any{int(100)},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "uint8 to string",
			dataArgs: []any{uint8(42)},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "uint16 to string",
			dataArgs: []any{uint16(1000)},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "int8 to string",
			dataArgs: []any{int8(50)},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "int16 to string",
			dataArgs: []any{int16(500)},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "int32 to string",
			dataArgs: []any{int32(10000)},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "uint to string",
			dataArgs: []any{uint(999)},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "float32 to string",
			dataArgs: []any{float32(3.14)},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "float64 to string",
			dataArgs: []any{float64(2.718)},
			dataRule: "s",
			wantErr:  false,
		},

		// uint8 ('b') tests
		{
			name:     "string to uint8",
			dataArgs: []any{"42"},
			dataRule: "b",
			wantErr:  false,
		},
		{
			name:     "uint8 to uint8",
			dataArgs: []any{uint8(100)},
			dataRule: "b",
			wantErr:  false,
		},
		{
			name:     "uint16 to uint8",
			dataArgs: []any{uint16(200)},
			dataRule: "b",
			wantErr:  false,
		},
		{
			name:     "uint16 overflow uint8",
			dataArgs: []any{uint16(300)},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "int16 to uint8",
			dataArgs: []any{int16(100)},
			dataRule: "b",
			wantErr:  false,
		},
		{
			name:     "int16 negative to uint8",
			dataArgs: []any{int16(-1)},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "int16 overflow uint8",
			dataArgs: []any{int16(300)},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "uint to uint8",
			dataArgs: []any{uint(50)},
			dataRule: "b",
			wantErr:  false,
		},
		{
			name:     "uint overflow uint8",
			dataArgs: []any{uint(500)},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "int to uint8",
			dataArgs: []any{int(100)},
			dataRule: "b",
			wantErr:  false,
		},
		{
			name:     "int negative to uint8",
			dataArgs: []any{int(-5)},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "int overflow uint8",
			dataArgs: []any{int(300)},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "float64 to uint8",
			dataArgs: []any{float64(50.5)},
			dataRule: "b",
			wantErr:  false,
		},
		{
			name:     "float64 negative to uint8",
			dataArgs: []any{float64(-1.0)},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "float64 overflow uint8",
			dataArgs: []any{float64(300.0)},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "invalid string to uint8",
			dataArgs: []any{"abc"},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "string overflow uint8",
			dataArgs: []any{"300"},
			dataRule: "b",
			wantErr:  true,
		},

		// uint16 ('w') tests
		{
			name:     "uint16",
			dataArgs: []any{uint16(1)},
			dataRule: "w",
			wantErr:  false,
		},
		{
			name:     "uint8 to uint16",
			dataArgs: []any{uint8(100)},
			dataRule: "w",
			wantErr:  false,
		},
		{
			name:     "int16 to uint16",
			dataArgs: []any{int16(1000)},
			dataRule: "w",
			wantErr:  false,
		},
		{
			name:     "int16 negative to uint16",
			dataArgs: []any{int16(-1)},
			dataRule: "w",
			wantErr:  true,
		},
		{
			name:     "uint to uint16",
			dataArgs: []any{uint(30000)},
			dataRule: "w",
			wantErr:  false,
		},
		{
			name:     "uint overflow uint16",
			dataArgs: []any{uint(70000)},
			dataRule: "w",
			wantErr:  true,
		},
		{
			name:     "int to uint16",
			dataArgs: []any{int(3)},
			dataRule: "w",
			wantErr:  false,
		},
		{
			name:     "int negative to uint16",
			dataArgs: []any{int(-10)},
			dataRule: "w",
			wantErr:  true,
		},
		{
			name:     "int overflow uint16",
			dataArgs: []any{int(70000)},
			dataRule: "w",
			wantErr:  true,
		},
		{
			name:     "float64 to uint16",
			dataArgs: []any{float64(2)},
			dataRule: "w",
			wantErr:  false,
		},
		{
			name:     "float64 negative to uint16",
			dataArgs: []any{float64(-5.0)},
			dataRule: "w",
			wantErr:  true,
		},
		{
			name:     "float64 overflow uint16",
			dataArgs: []any{float64(70000.0)},
			dataRule: "w",
			wantErr:  true,
		},
		{
			name:     "string to uint16",
			dataArgs: []any{"111"},
			dataRule: "w",
			wantErr:  false,
		},
		{
			name:     "invalid string to uint16",
			dataArgs: []any{"one"},
			dataRule: "w",
			wantErr:  true,
		},
		{
			name:     "string overflow uint16",
			dataArgs: []any{"70000"},
			dataRule: "w",
			wantErr:  true,
		},

		// Error cases
		{
			name:     "wrong number of args",
			dataArgs: []any{"test", uint16(1)},
			dataRule: "s",
			wantErr:  true,
		},
		{
			name:     "unknown arg type",
			dataArgs: []any{uint32(100)},
			dataRule: "d",
			wantErr:  true,
		},
		{
			name:     "unsupported type to uint8",
			dataArgs: []any{int32(50)},
			dataRule: "b",
			wantErr:  true,
		},
		{
			name:     "unsupported type to uint16",
			dataArgs: []any{int32(50)},
			dataRule: "w",
			wantErr:  true,
		},

		// Multiple args
		{
			name:     "multiple args mixed",
			dataArgs: []any{"test", uint8(42), uint16(1000)},
			dataRule: "sbw",
			wantErr:  false,
		},
		{
			name:     "multiple args with conversions",
			dataArgs: []any{int(100), "255", float64(50.0)},
			dataRule: "sbb",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PaveArgs(tt.dataArgs, tt.dataRule); (err != nil) != tt.wantErr {
				t.Errorf("PaveArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
