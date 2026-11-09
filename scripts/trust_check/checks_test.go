package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func copyTrustFiles(t *testing.T, dst string) {
	t.Helper()
	for _, f := range []string{"claims.md", "evidence-manifest.json", "verdict.json", "limitations.md"} {
		data, err := os.ReadFile(filepath.Join(repoRoot(t), "trust", f))
		if err != nil {
			t.Fatalf("fixture source missing: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(dst, "trust"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, "trust", f), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// editManifestJSON / editVerdictJSON mutate a fixture file in place.
func editJSON(t *testing.T, path string, fn func(m map[string]interface{})) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	fn(m)
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func overallOf(t *testing.T, root string) (*Report, string) {
	t.Helper()
	rep, err := runChecks(root)
	if err != nil {
		t.Fatal(err)
	}
	return rep, rep.Overall
}

func failCheck(t *testing.T, rep *Report, id string) {
	t.Helper()
	for _, c := range rep.Checks {
		if c.ID == id && c.Status == "FAIL" {
			return
		}
	}
	t.Errorf("expected FAIL on check %s; got report overall=%s checks=%+v", id, rep.Overall, rep.Checks)
}

// The live repository record must be mechanically green.
func TestCleanRecordPasses(t *testing.T) {
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, "trust", "verdict.json")); err != nil {
		t.Skip("no trust/ record in repo")
	}
	rep, overall := overallOf(t, root)
	if overall != "PASS" {
		for _, c := range rep.Checks {
			if c.Status == "FAIL" {
				t.Errorf("check %s FAILED: %s", c.ID, c.Detail)
			}
		}
		t.Fatalf("expected overall PASS, got %s", overall)
	}
}

// Negative control 1: tampered artifact hash must break the report.
func TestBrokenHashFails(t *testing.T) {
	dir := t.TempDir()
	copyTrustFiles(t, dir)
	man := filepath.Join(dir, "trust", "evidence-manifest.json")
	editJSON(t, man, func(m map[string]interface{}) {
		ev, _ := m["evidence"].([]interface{})
		first, _ := ev[0].(map[string]interface{})
		if h, ok := first["sha256"].(string); ok && h != "" {
			flip := "0"
			if h[0] == '0' {
				flip = "1"
			}
			first["sha256"] = flip + h[1:]
		}
	})
	rep, overall := overallOf(t, dir)
	if overall != "FAIL" {
		t.Fatalf("expected overall FAIL, got %s", overall)
	}
	failCheck(t, rep, "HASHES_MATCH")
}

// Negative control 2: INCONCLUSIVE without reason_code must break the report.
func TestMissingReasonCodeFails(t *testing.T) {
	dir := t.TempDir()
	copyTrustFiles(t, dir)
	ver := filepath.Join(dir, "trust", "verdict.json")
	editJSON(t, ver, func(m map[string]interface{}) {
		claims, _ := m["claims"].([]interface{})
		for _, c := range claims {
			cm, _ := c.(map[string]interface{})
			if cm["status"] == "INCONCLUSIVE" {
				delete(cm, "reason_code")
			}
		}
	})
	rep, overall := overallOf(t, dir)
	if overall != "FAIL" {
		t.Fatalf("expected overall FAIL, got %s", overall)
	}
	failCheck(t, rep, "REASON_CODES")
}

// Negative control 3: statistical PASS without a negative control must fail.
func TestPASSWithoutControlFails(t *testing.T) {
	dir := t.TempDir()
	copyTrustFiles(t, dir)
	ver := filepath.Join(dir, "trust", "verdict.json")
	editJSON(t, ver, func(m map[string]interface{}) {
		claims, _ := m["claims"].([]interface{})
		for _, c := range claims {
			cm, _ := c.(map[string]interface{})
			if cm["control"] == "NC-NOISE" {
				delete(cm, "control")
			}
		}
	})
	rep, overall := overallOf(t, dir)
	if overall != "FAIL" {
		t.Fatalf("expected overall FAIL, got %s", overall)
	}
	failCheck(t, rep, "CONTROLS")
}

// I-001 regression: a criteria-lock commit carrying an EMPTY claims.md
// must be reported FAIL, mirroring the GitHub-incident pattern.
func TestLockWithEmptyBlobFails(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "trust"), 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) string {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("init", "-q", "-b", "main")

	if err := os.WriteFile(filepath.Join(dir, "trust", "claims.md"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	run("-c", "user.email=t@local", "-c", "user.name=fixtures", "add", "-A")
	run("-c", "user.email=t@local", "-c", "user.name=fixtures", "commit", "-q", "-m", "lock (empty claims — the I-001 defect)")
	lockCommit := run("rev-parse", "HEAD")

	ver := map[string]interface{}{
		"verdict_id": "V-FIXTURE", "date": "2026-11-09", "valid_until": "object-change",
		"object_version": "fixture",
		"claims": []map[string]interface{}{{
			"id": "C-01", "status": "PASS", "verification": "V-01: clean run", "evidence": []string{"E-01"},
		}},
	}
	man := map[string]interface{}{
		"schema": "trust-manifest/v3.5", "date": "2026-11-09",
		"criteria_lock": map[string]interface{}{"commit": lockCommit, "date": "2026-11-09", "note": "empty-blob lock fixture"},
		"evidence": []map[string]interface{}{{
			"id": "E-01", "type": "code", "path": "trust/claims.md", "sha256": nil,
		}},
		"controls": []map[string]interface{}{{"id": "NC-FIXTURE", "N": 50}},
	}
	writeJSON := func(name string, v interface{}) {
		b, _ := json.MarshalIndent(v, "", "  ")
		if err := os.WriteFile(filepath.Join(dir, "trust", name), append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeJSON("evidence-manifest.json", man)
	writeJSON("verdict.json", ver)
	if err := os.WriteFile(filepath.Join(dir, "trust", "limitations.md"), []byte("# fixture"), 0o644); err != nil {
		t.Fatal(err)
	}

	rep, _ := overallOf(t, dir)
	failCheck(t, rep, "LOCK_BLOB")
}
