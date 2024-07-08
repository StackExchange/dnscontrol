package rtypecontrol

import "testing"

func TestValidateArgs(t *testing.T) {
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
			wantErr:  true,
		},

		{
			name:     "int",
			dataArgs: []any{int(1)},
			dataRule: "i",
			wantErr:  false,
		},
		{
			name:     "string to int",
			dataArgs: []any{"111"},
			dataRule: "i",
			wantErr:  false,
		},
		{
			name:     "txt to int",
			dataArgs: []any{"one"},
			dataRule: "i",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckArgTypes(tt.dataArgs, tt.dataRule); (err != nil) != tt.wantErr {
				t.Errorf("ValidateArgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
