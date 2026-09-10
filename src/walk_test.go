package walkcontent

import "testing"

func TestExtractDate(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantDate string
		wantOk   bool
	}{
		{
			name:     "date in filename",
			path:     "posts/2024-01-15-hello.md",
			wantDate: "2024-01-15",
			wantOk:   true,
		},
		{
			name:     "date as directory segment",
			path:     "posts/2024-01-15/index.md",
			wantDate: "2024-01-15",
			wantOk:   true,
		},
		{
			name:     "YYYY/MM-DD",
			path:     "posts/2024/01-15-hello.md",
			wantDate: "2024-01-15",
			wantOk:   true,
		},
		{
			name:     "YYYY/MM/DD",
			path:     "posts/2024/01/15/hello.md",
			wantDate: "2024-01-15",
			wantOk:   true,
		},
		{
			name:   "no date",
			path:   "posts/hello.md",
			wantOk: false,
		},
		{
			name:   "invalid calendar date is rejected",
			path:   "posts/2024-13-40-hello.md",
			wantOk: false,
		},
		{
			name:     "first match wins",
			path:     "2024-01-15/archive-2023-06-01.md",
			wantDate: "2024-01-15",
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := extractDate(tt.path)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && got != tt.wantDate {
				t.Fatalf("date = %q, want %q", got, tt.wantDate)
			}
		})
	}
}
