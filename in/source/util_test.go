package source

import (
	"bytes"
	"testing"
)

func TestFill0(t *testing.T) {
	tests := []struct {
		name      string
		message   []byte
		chunkSize int
		expected  []byte
	}{
		{
			name:      "length equal to chunkSize",
			message:   []byte{1, 2, 3, 4},
			chunkSize: 4,
			expected:  []byte{1, 2, 3, 4},
		},
		{
			name:      "length less than chunkSize",
			message:   []byte{1, 2},
			chunkSize: 4,
			expected:  []byte{1, 2, 0, 0},
		},
		{
			name:      "length greater than chunkSize and not a multiple",
			message:   []byte{1, 2, 3, 4, 5},
			chunkSize: 4,
			expected:  []byte{1, 2, 3, 4, 5, 0, 0, 0},
		},
		{
			name:      "empty slice, with chunkSize",
			message:   []byte{},
			chunkSize: 4,
			expected:  []byte{0, 0, 0, 0},
		},
		{
			name:      "non-standard chunkSize",
			message:   []byte{1, 2, 3},
			chunkSize: 2,
			expected:  []byte{1, 2, 3, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Fill0(tt.message, tt.chunkSize)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("fill0(%v, %d) = %v, want %v", tt.message, tt.chunkSize, result, tt.expected)
			}
		})
	}
}
