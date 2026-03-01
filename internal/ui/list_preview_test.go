package ui

import "testing"

func TestPreviewForQueryVariants(t *testing.T) {
	if got := previewForQuery("", "x"); got != "(empty)" {
		t.Fatalf("unexpected empty preview: %s", got)
	}
	long := "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789"
	if got := previewForQuery(long, ""); len(got) == 0 {
		t.Fatalf("expected non-empty preview")
	}
	withHit := "first line contains needle right here and then continues"
	got := previewForQuery(withHit, "needle")
	if got == "" {
		t.Fatalf("expected preview with match")
	}
	if got = previewForQuery(withHit, "missing"); got == "" {
		t.Fatalf("expected fallback preview")
	}
}
