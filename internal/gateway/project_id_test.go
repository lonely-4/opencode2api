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
	oldSum := sha256.Sum256([]byte("opencode-project\x00" + project))
	old := hex.EncodeToString(oldSum[:20])
	if got == old {
		t.Fatalf("v2 domain still matches the previous project id %s", got)
	}
	if projectID("global") != "global" {
		t.Fatal("global project was rewritten")
	}
	commit := "0123456789abcdef0123456789abcdef01234567"
	if projectID(commit) != commit {
		t.Fatal("git commit project was rewritten")
	}
}
