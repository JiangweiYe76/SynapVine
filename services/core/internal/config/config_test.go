package config

import (
	"reflect"
	"testing"
)

func TestParseServiceTokens(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want map[string]string
	}{
		{
			name: "full configuration",
			raw:  "portal=t1,console=t2,discovery=t3",
			want: map[string]string{"portal": "t1", "console": "t2", "discovery": "t3"},
		},
		{
			name: "surrounding whitespace is trimmed",
			raw:  " portal = t1 , console = t2 ",
			want: map[string]string{"portal": "t1", "console": "t2"},
		},
		{
			name: "trailing comma is ignored",
			raw:  "portal=t1,console=t2,",
			want: map[string]string{"portal": "t1", "console": "t2"},
		},
		{
			name: "empty value is skipped",
			raw:  "portal=t1,console=",
			want: map[string]string{"portal": "t1"},
		},
		{
			name: "empty name is skipped",
			raw:  "portal=t1,=t2",
			want: map[string]string{"portal": "t1"},
		},
		{
			name: "entry without separator is skipped",
			raw:  "portal=t1,garbage",
			want: map[string]string{"portal": "t1"},
		},
		{
			name: "empty string yields empty map",
			raw:  "",
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseServiceTokens(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseServiceTokens(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
