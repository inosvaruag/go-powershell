package powershell

import (
	"strings"
	"testing"
)

// TestStreamReaderExtractsInputBeforeBoundary tests that streamReader extracts the input stream
// before boundary correctly.
func TestStreamReaderExtractsInputBeforeBoundary(t *testing.T) {
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

// TestStreamReaderIgnoresCharsAfterBoundary ensures streamReader() ignores extraneous
// characters after the boundary.
func TestStreamReaderIgnoresCharsAfterBoundary(t *testing.T) {
	stream := "abcdefghijklmnopqrstuvwxyz"
	boundary := "Z"
	extraChars := "garbage"
	totalStream := stream + boundary + newline + extraChars
	reader := strings.NewReader(totalStream)
	output := ""
	streamReader(reader, boundary, &output, "test")
	if output != stream {
		t.Errorf("stream not read correctly, got: %v, want: %v", output, stream)
	}
}

// TestStreamReaderReturnsErrorIfEOFIsEncounteredBeforeBoundary ensures streamReader() returns
// an error if EOF is returned before boundary is read
func TestStreamReaderReturnsErrorIfEOFIsEncounteredBeforeBoundary(t *testing.T) {
	stream := "abcdefghijklmnopqrstuvwxyz"
	boundary := "Z"
	reader := strings.NewReader(stream)
	output := ""
	err := streamReader(reader, boundary, &output, "test")
	if err == nil {
		t.Errorf("no error returned when EOF is reached before boundary, got: nil. want: non-nil")
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
