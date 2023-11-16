package random

import (
	"testing"
)

func TestNewRandomString(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{
			name: "size = 1",
			size: 1,
		},
		{
			name: "size = 5",
			size: 5,
		},
		{
			name: "size = 10",
			size: 10,
		},
		{
			name: "size = 20",
			size: 20,
		},
		{
			name: "size = 30",
			size: 30,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str1 := NewRandomString(tt.size)
			str2 := NewRandomString(tt.size)

			if len(str1) != tt.size {
				t.Errorf("Len of NewRandomString: %d, expected len: %d", len(str1), tt.size)
			}

			if len(str2) != tt.size {
				t.Errorf("Len of NewRandomString: %d, expected len: %d", len(str1), tt.size)
			}

			if str1 == str2 {
				t.Errorf("random generated equal strings: %s", str1)
			}
		})
	}
}
