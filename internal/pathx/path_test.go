package pathx

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    []string
		wantErr bool
	}{
		{
			name: "simple path",
			path: "version",
			want: []string{"version"},
		},
		{
			name: "path with leading dot",
			path: ".version",
			want: []string{"version"},
		},
		{
			name: "nested path",
			path: ".app.version",
			want: []string{"app", "version"},
		},
		{
			name: "deeply nested",
			path: "a.b.c.d",
			want: []string{"a", "b", "c", "d"},
		},
		{
			name:    "empty",
			path:    "",
			wantErr: true,
		},
		{
			name:    "just dot",
			path:    ".",
			wantErr: true,
		},
		{
			name:    "empty component",
			path:    "a..b",
			wantErr: true,
		},
		{
			name:    "trailing dot",
			path:    "a.b.",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("Parse() = %v, want %v", got, tt.want)
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("Parse()[%d] = %v, want %v", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name    string
		data    any
		path    []string
		want    any
		wantErr bool
	}{
		{
			name: "get from map",
			data: map[string]any{"version": "1.0.0"},
			path: []string{"version"},
			want: "1.0.0",
		},
		{
			name: "get nested",
			data: map[string]any{"app": map[string]any{"version": "1.0.0"}},
			path: []string{"app", "version"},
			want: "1.0.0",
		},
		{
			name:    "path not found",
			data:    map[string]any{"version": "1.0.0"},
			path:    []string{"missing"},
			wantErr: true,
		},
		{
			name:    "cannot traverse",
			data:    map[string]any{"version": "1.0.0"},
			path:    []string{"version", "nested"},
			wantErr: true,
		},
		{
			name: "interface map",
			data: map[any]any{"version": "1.0.0"},
			path: []string{"version"},
			want: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Get(tt.data, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name    string
		data    any
		path    []string
		value   any
		wantErr bool
	}{
		{
			name:  "set in map",
			data:  map[string]any{"version": "1.0.0"},
			path:  []string{"version"},
			value: "2.0.0",
		},
		{
			name:  "set nested",
			data:  map[string]any{"app": map[string]any{"version": "1.0.0"}},
			path:  []string{"app", "version"},
			value: "2.0.0",
		},
		{
			name:    "path not found",
			data:    map[string]any{"version": "1.0.0"},
			path:    []string{"missing"},
			value:   "2.0.0",
			wantErr: true,
		},
		{
			name:    "empty path",
			data:    map[string]any{"version": "1.0.0"},
			path:    []string{},
			value:   "2.0.0",
			wantErr: true,
		},
		{
			name:    "cannot traverse",
			data:    map[string]any{"version": "1.0.0"},
			path:    []string{"version", "nested"},
			value:   "2.0.0",
			wantErr: true,
		},
		{
			name:  "interface map",
			data:  map[any]any{"version": "1.0.0"},
			path:  []string{"version"},
			value: "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Set(tt.data, tt.path, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{
			name:  "single",
			parts: []string{"version"},
			want:  "version",
		},
		{
			name:  "multiple",
			parts: []string{"app", "version"},
			want:  "app.version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Join(tt.parts); got != tt.want {
				t.Errorf("Join() = %v, want %v", got, tt.want)
			}
		})
	}
}
