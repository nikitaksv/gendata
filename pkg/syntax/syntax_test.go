package syntax

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		in      []byte
		wantErr bool
	}{
		{
			name:    "valid Properties",
			wantErr: false,
			in:      []byte(`{{Properties}}{{Name.CamelCase}}{{/Properties}}`),
		},
		{
			name:    "valid type.IsArray",
			wantErr: false,
			in:      []byte(`{{Type.IsArray}}{{Type}}[]{{/Type.IsArray}}`),
		},
		{
			name:    "invalid Properties not close",
			wantErr: true,
			in:      []byte(`{{Properties}}{{Name.CamelCase}}`),
		},
		{
			name:    "invalid Properties not open",
			wantErr: true,
			in:      []byte(`{{Name.CamelCase}}{{/Properties}}`),
		},
		{
			name:    "invalid Type.IsArray not close",
			wantErr: true,
			in:      []byte(`{{Type.IsArray}}{{Type}}[]`),
		},
		{
			name:    "invalid Type.IsArray not open",
			wantErr: true,
			in:      []byte(`{{Type}}[]{{Type.IsArray}}`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := Validate(test.in); err != nil && !test.wantErr {
				t.Error(err)
			}
		})
	}
}
