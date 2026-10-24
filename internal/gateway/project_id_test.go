package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestProjectIDDomainSeparatesVersions(t *testing.T) {
	const project = "opencode2api:default-project:server-key"
	got := projectID(project)
	if len(got) != 40 {
		t.Fatalf("project id length = %d, want 40", len(got))
	}
	for _, domain := range []string{"opencode-project\x00", "opencode-project/v2\x00"} {
		sum := sha256.Sum256([]byte(domain + project))
		if got == hex.EncodeToString(sum[:20]) {
			t.Fatalf("v3 domain still matches %q", domain)
		}
	}
	if projectID("global") != "global" {
		t.Fatal("global project was rewritten")
	}
	commit := "0123456789abcdef0123456789abcdef01234567"
	rewritten := projectID(commit)
	if rewritten == commit || len(rewritten) != 40 {
		t.Fatalf("git commit project = %s, want a new 40-hex id", rewritten)
	}
}
