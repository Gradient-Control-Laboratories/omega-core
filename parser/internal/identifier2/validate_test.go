package identifier2

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		ident   string
		wantErr bool
	}{
		{
			name:    "valid identifier",
			ident:   "my_var",
			wantErr: false,
		},
		{
			name:    "reserved word 'if'",
			ident:   "if",
			wantErr: true,
		},
		{
			name:    "reserved word 'const'",
			ident:   "const",
			wantErr: true,
		},
		{
			name:    "reserved keyword 'auto'",
			ident:   "auto",
			wantErr: true,
		},
		{
			name:    "builtin 'clamp'",
			ident:   "clamp",
			wantErr: true,
		},
		{
			name:    "empty",
			ident:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.ident)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tt.ident, err, tt.wantErr)
			}
		})
	}
}
