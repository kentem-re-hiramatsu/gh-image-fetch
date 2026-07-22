package attachment

import "testing"

func TestParse(t *testing.T) {
	const id = "12345678-abcd-4ef0-9876-0123456789ab"

	tests := []struct {
		name    string
		input   string
		want    Asset
		wantErr bool
	}{
		{
			name:  "bare uuid",
			input: id,
			want:  Asset{Host: "github.com", ID: id},
		},
		{
			name:  "full url",
			input: "https://github.com/user-attachments/assets/" + id,
			want:  Asset{Host: "github.com", ID: id},
		},
		{
			name:  "full url with trailing slash",
			input: "https://github.com/user-attachments/assets/" + id + "/",
			want:  Asset{Host: "github.com", ID: id},
		},
		{
			name:  "uppercase host",
			input: "https://GitHub.com/user-attachments/assets/" + id,
			want:  Asset{Host: "github.com", ID: id},
		},
		{name: "empty", input: "", wantErr: true},
		{name: "whitespace", input: "   ", wantErr: true},
		{name: "not a uuid or url", input: "hello", wantErr: true},
		{name: "http scheme", input: "http://github.com/user-attachments/assets/" + id, wantErr: true},
		{name: "wrong host", input: "https://example.com/user-attachments/assets/" + id, wantErr: true},
		{name: "wrong path", input: "https://github.com/foo/assets/" + id, wantErr: true},
		{name: "malformed uuid in url", input: "https://github.com/user-attachments/assets/not-a-uuid", wantErr: true},
		{name: "release asset url", input: "https://github.com/owner/repo/releases/download/v1/a.png", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) = %+v, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("Parse(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestAssetURL(t *testing.T) {
	a := Asset{Host: "github.com", ID: "12345678-abcd-4ef0-9876-0123456789ab"}
	want := "https://github.com/user-attachments/assets/12345678-abcd-4ef0-9876-0123456789ab"
	if got := a.URL(); got != want {
		t.Fatalf("URL() = %q, want %q", got, want)
	}
}
