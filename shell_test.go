package powershell

import (
	"strings"
	"testing"
)

// TestStreamReader tests that streamReader extracts the input stream before boundary correctly.
func TestStreamReader(t *testing.T) {
	stream := "abcdefghijklmnopqrstuvwxyz"
	boundary := "Z"
	totalStream := stream + boundary + newline
	reader := strings.NewReader(totalStream)
	output := ""
	streamReader(reader, boundary, &output, "test")
	if output != stream {
		t.Errorf("stream not read correctly, got: %v, want: %v", output, stream)
	}
}

// BenchmarkStreamReader benchmarks streamReader performance.
func BenchmarkStreamReader(b *testing.B) {
	stream := strings.Repeat("x", b.N)
	boundary := "a"
	totalStream := stream + boundary + newline
	output := ""
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		reader := strings.NewReader(totalStream)
		streamReader(reader, boundary, &output, "test")
	}

	b.StopTimer()
	if output != stream {
		b.Errorf("unexpected result; got=%s, want=%s", output, stream)
	}
}
