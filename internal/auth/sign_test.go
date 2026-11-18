package auth

import "testing"

func TestSignMatchesPython(t *testing.T) {
	// Computed in Python:
	// hmac.new(b"test-token", "abc|1|LETTER|É".encode("utf-8"), hashlib.sha256).hexdigest()
	want := "67b2b1f992d7eb7160b5d31c7db53cc80034a9b70c906cb76519a7c2ba379c73"
	got := Sign("test-token", "abc|1|LETTER|É")
	if got != want {
		t.Errorf("got %q , want %q", got, want)
	}
}
