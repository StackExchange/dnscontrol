package rtypecontrol

import "testing"

func TestPaveArgs(t *testing.T) {
	tests := []struct {
		name     string
		dataArgs []any
		dataRule string
		wantErr  bool
	}{
		{
			name:     "string",
			dataArgs: []any{"one"},
			dataRule: "s",
			wantErr:  false,
		},
		{
			name:     "int to string",
			dataArgs: []any{100},
			dataRule: "s",
			wantErr:  false,
		},

		{
			name:     "uint16",
			dataArgs: []any{uint16(1)},
			dataRule: "i",
			wantErr:  false,
		},
		{
			name:     "float to uint16",
			dataArgs: []any{float64(2)},
			dataRule: "i",
			wantErr:  false,
		},
		{
			name:     "int uint16",
			dataArgs: []any{int(3)},
			dataRule: "i",
			wantErr:  false,
		},
		{
			name:     "string to uint16",
			dataArgs: []any{"111"},
			dataRule: "i",
			wantErr:  false,
		},
		{
			name:     "txt to uint16",
			dataArgs: []any{"one"},
			dataRule: "i",
			wantErr:  true,
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
