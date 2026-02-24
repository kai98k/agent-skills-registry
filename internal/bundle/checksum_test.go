package bundle

import (
	"testing"
)

func TestSHA256Bytes(t *testing.T) {
	hash := SHA256Bytes([]byte("hello world"))
	expected := "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
	if hash != expected {
		t.Errorf("SHA256Bytes = %q, want %q", hash, expected)
	}
}

func TestSHA256Bytes_Empty(t *testing.T) {
	hash := SHA256Bytes([]byte{})
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expected {
		t.Errorf("SHA256Bytes empty = %q, want %q", hash, expected)
	}
}
