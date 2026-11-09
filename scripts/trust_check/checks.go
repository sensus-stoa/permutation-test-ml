package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const toolName = "trust_check"

type Check struct {
	ID     string `json:"id"`
	Status string `json:"status"` // PASS | FAIL | SKIP
	Detail string `json:"detail"`
}

type ClaimRef struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Verification string `json:"verification,omitempty"`
	Control      string `json:"control,omitempty"`
}

type Report struct {
	Tool      string     `json:"tool"`
	Version   string     `json:"version"`
	CheckedAt string     `json:"checked_at"`
	Root      string     `json:"root"`
	Overall   string     `json:"overall"` // PASS | FAIL
	Claims    []ClaimRef `json:"claims"`
	Checks    []Check    `json:"checks"`
}

type manifestFile struct {
	CriteriaLock struct {
		Commit string `json:"commit"`
		Date   string `json:"date"`
		Note   string `json:"note"`
	} `json:"criteria_lock"`
	Evidence []struct {
		ID     string  `json:"id"`
		Type   string  `json:"type"`
		Path   string  `json:"path"`
		SHA256 *string `json:"sha256"`
	} `json:"evidence"`
	Controls []struct {
		ID string `json:"id"`
		N  int    `json:"N"`
	} `json:"controls"`
}

type verdictFile struct {
	VerdictID     string `json:"verdict_id"`
	Date          string `json:"date"`
	ValidUntil    string `json:"valid_until"`
	ObjectVersion string `json:"object_version"`
	Claims        []struct {
		ID           string   `json:"id"`
		Status       string   `json:"status"`
		Verification string   `json:"verification"`
		Evidence     []string `json:"evidence"`
		Control      string   `json:"control"`
		ReasonCode   string   `json:"reason_code"`
	} `json:"claims"`
}

var claimHeading = regexp.MustCompile(`(?m)^##\s+(C-\d+)`)

func writeReport(path string, rep *Report) error {
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

type reportBuilder struct{ checks []Check }

func (r *reportBuilder) add(id, status, detail string) {
	r.checks = append(r.checks, Check{ID: id, Status: status, Detail: detail})
}

func (r *reportBuilder) failIf(cond bool, id, okDetail, failDetail string) {
	if cond {
		r.add(id, "FAIL", failDetail)
	} else {
		r.add(id, "PASS", okDetail)
	}
}

// runChecks validates the mechanical integrity of a Trust Record located at
// <root>/trust/. It never judges meaning: a green report states "the
// mechanics hold", nothing more.
func runChecks(root string) (*Report, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	rep := &Report{
		Tool:      toolName,
		Version:   version,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Root:      abs,
	}
	r := &reportBuilder{}
	trustDir := filepath.Join(abs, "trust")

	// 1. presence & non-emptiness of the four record files
	files := []string{"claims.md", "evidence-manifest.json", "verdict.json", "limitations.md"}
	var missing []string
	for _, f := range files {
		fi, err := os.Stat(filepath.Join(trustDir, f))
		if err != nil || fi.Size() == 0 {
			missing = append(missing, f)
		}
	}
	r.failIf(len(missing) > 0, "FILES_PRESENT",
		fmt.Sprintf("all %d record files present and non-empty", len(files)),
		"missing/empty: "+strings.Join(missing, ", "))

	// 2. JSON validity and minimal structure
	manBytes, errMan := os.ReadFile(filepath.Join(trustDir, "evidence-manifest.json"))
	verBytes, errVer := os.ReadFile(filepath.Join(trustDir, "verdict.json"))
	var man manifestFile
	var ver verdictFile
	jsonOk := true
	jsonState := "manifest and verdict parse; structure complete"
	switch {
	case errMan != nil:
		jsonOk, jsonState = false, "evidence-manifest.json unreadable: "+errMan.Error()
	case errVer != nil:
		jsonOk, jsonState = false, "verdict.json unreadable: "+errVer.Error()
	default:
		if e := json.Unmarshal(manBytes, &man); e != nil {
			jsonOk, jsonState = false, "evidence-manifest.json invalid: "+e.Error()
		} else if e := json.Unmarshal(verBytes, &ver); e != nil {
			jsonOk, jsonState = false, "verdict.json invalid: "+e.Error()
		} else if len(man.Evidence) == 0 {
			jsonOk, jsonState = false, "manifest lists no evidence"
		} else if len(ver.Claims) == 0 {
			jsonOk, jsonState = false, "verdict lists no claims"
		} else if strings.TrimSpace(man.CriteriaLock.Commit) == "" {
			jsonOk, jsonState = false, "criteria_lock.commit missing"
		}
	}
	r.failIf(!jsonOk, "JSON_VALID", jsonState, jsonState)

	if jsonOk {
		// 3. sha256 of hashed evidence artifacts
		var bad, skipped []string
		verified := 0
		for _, ev := range man.Evidence {
			if ev.Path == "" || ev.SHA256 == nil || strings.TrimSpace(*ev.SHA256) == "" {
				skipped = append(skipped, ev.ID)
				continue
			}
			data, err := os.ReadFile(filepath.Join(abs, filepath.FromSlash(ev.Path)))
			if err != nil {
				bad = append(bad, fmt.Sprintf("%s: unreadable: %v", ev.ID, err))
				continue
			}
			sum := sha256.Sum256(data)
			got := hex.EncodeToString(sum[:])
			if !strings.EqualFold(got, strings.TrimSpace(*ev.SHA256)) {
				bad = append(bad, fmt.Sprintf("%s: declared %s..., actual %s...", ev.ID, shortHash(*ev.SHA256), shortHash(got)))
				continue
			}
			verified++
		}
		status, detail := "PASS", fmt.Sprintf("%d hashed artifacts verified", verified)
		if len(skipped) > 0 {
			detail += fmt.Sprintf("; %d skipped (no hash by design: %s)", len(skipped), strings.Join(skipped, ","))
		}
		if len(bad) > 0 {
			status, detail = "FAIL", detail+"; MISMATCH: "+strings.Join(bad, "; ")
		}
		r.add("HASHES_MATCH", status, detail)

		// 4. every verdict claim id exists in claims.md
		claimsMD, err := os.ReadFile(filepath.Join(trustDir, "claims.md"))
		if err != nil {
			r.add("CLAIMS_XREF", "FAIL", "claims.md unreadable: "+err.Error())
		} else {
			known := map[string]bool{}
			for _, m := range claimHeading.FindAllStringSubmatch(string(claimsMD), -1) {
				known[m[1]] = true
			}
			var absent []string
			for _, c := range ver.Claims {
				if !known[c.ID] {
					absent = append(absent, c.ID)
				}
			}
			r.failIf(len(absent) > 0, "CLAIMS_XREF",
				fmt.Sprintf("all %d claim ids present in claims.md", len(ver.Claims)),
				"in verdict but absent in claims.md: "+strings.Join(absent, ", "))
		}

		// 5. every evidence id referenced by verdict exists in manifest
		evIDs := map[string]bool{}
		for _, ev := range man.Evidence {
			evIDs[ev.ID] = true
		}
		var dangling []string
		for _, c := range ver.Claims {
			for _, e := range c.Evidence {
				if !evIDs[e] {
					dangling = append(dangling, c.ID+"->"+e)
				}
			}
		}
		r.failIf(len(dangling) > 0, "EVIDENCE_XREF",
			"all evidence references resolve to manifest entries",
			"dangling evidence references: "+strings.Join(dangling, ", "))

		// 6. statistical PASS claims carry a negative control with N >= 1
		ctrl := map[string]int{}
		for _, c := range man.Controls {
			ctrl[c.ID] = c.N
		}
		var ctrlIssues []string
		for _, c := range ver.Claims {
			if c.Status != "PASS" {
				continue
			}
			if c.Control != "" {
				n, ok := ctrl[c.Control]
				if !ok {
					ctrlIssues = append(ctrlIssues, c.ID+": control "+c.Control+" not defined in manifest")
				} else if n < 1 {
					ctrlIssues = append(ctrlIssues, fmt.Sprintf("%s: control %s has N < 1", c.ID, c.Control))
				}
			} else if strings.Contains(c.Verification, "V-02") {
				ctrlIssues = append(ctrlIssues, c.ID+": statistical PASS (verification references V-02) without negative control")
			}
		}
		r.failIf(len(ctrlIssues) > 0, "CONTROLS",
			fmt.Sprintf("controls sound for %d PASS claims", countPASS(ver.Claims)),
			strings.Join(ctrlIssues, "; "))

		// 7. FAIL / INCONCLUSIVE claims carry a reason_code
		var noReason []string
		for _, c := range ver.Claims {
			if (c.Status == "FAIL" || c.Status == "INCONCLUSIVE") && strings.TrimSpace(c.ReasonCode) == "" {
				noReason = append(noReason, c.ID)
			}
		}
		r.failIf(len(noReason) > 0, "REASON_CODES",
			"every non-PASS claim carries a reason_code",
			"FAIL/INCONCLUSIVE without reason_code: "+strings.Join(noReason, ", "))

		// 8. verdict validity horizon
		r.failIf(strings.TrimSpace(ver.ValidUntil) == "", "VALID_UNTIL",
			"valid_until: "+ver.ValidUntil,
			"verdict.valid_until is empty")

		for _, c := range ver.Claims {
			rep.Claims = append(rep.Claims, ClaimRef{
				ID: c.ID, Status: c.Status, Verification: c.Verification, Control: c.Control,
			})
		}
	}

	// 9. criteria-lock commit integrity (incident I-001 lesson)
	if isGitRepo(abs) && jsonOk {
		lock := man.CriteriaLock.Commit
		out, err := gitShowBytes(abs, lock, "trust/claims.md")
		if err != nil {
			r.add("LOCK_BLOB", "FAIL", "criteria-lock commit unreadable: "+short(err.Error()))
		} else if len(bytes.TrimSpace(out)) == 0 {
			r.add("LOCK_BLOB", "FAIL",
				"criteria-lock commit contains an EMPTY trust/claims.md — incident I-001 pattern; the lock is invalid")
		} else {
			r.add("LOCK_BLOB", "PASS",
				fmt.Sprintf("locked trust/claims.md non-empty (%d bytes at %s...)", len(out), shortHash(lock)))
		}
		if err := gitAncestor(abs, lock); err != nil {
			r.add("LOCK_ANCESTOR", "FAIL", "criteria-lock commit is not an ancestor of HEAD (lock from a side branch?)")
		} else {
			r.add("LOCK_ANCESTOR", "PASS", shortHash(lock)+"... is an ancestor of HEAD")
		}
	} else {
		r.add("LOCK_BLOB", "SKIP", "not a git checkout or manifest unreadable")
		r.add("LOCK_ANCESTOR", "SKIP", "not a git checkout or manifest unreadable")
	}

	rep.Checks = r.checks
	rep.Overall = "PASS"
	for _, c := range rep.Checks {
		if c.Status == "FAIL" {
			rep.Overall = "FAIL"
			break
		}
	}
	return rep, nil
}

func countPASS(claims []struct {
	ID           string   `json:"id"`
	Status       string   `json:"status"`
	Verification string   `json:"verification"`
	Evidence     []string `json:"evidence"`
	Control      string   `json:"control"`
	ReasonCode   string   `json:"reason_code"`
}) int {
	n := 0
	for _, c := range claims {
		if c.Status == "PASS" {
			n++
		}
	}
	return n
}

// isGitRepo checks that root itself is a git work tree (direct .git entry:
// a directory for normal clones, a file for worktrees/submodules). It
// deliberately does NOT walk up: a validator must never inherit a parent
// repository's git context (tmp dirs nested inside unrelated checkouts).
func isGitRepo(root string) bool {
	_, err := os.Stat(filepath.Join(root, ".git"))
	return err == nil
}

func git(r *exec.Cmd) (string, error) {
	var out, errb bytes.Buffer
	r.Stdout = &out
	r.Stderr = &errb
	if err := r.Run(); err != nil {
		if errb.Len() > 0 {
			return "", fmt.Errorf("%v: %s", err, strings.TrimSpace(errb.String()))
		}
		return "", err
	}
	return out.String(), nil
}

func gitShowBytes(root, commit, path string) ([]byte, error) {
	cmd := exec.Command("git", "-C", root, "show", commit+":"+path)
	var out, errb bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errb
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%v: %s", err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}

func gitAncestor(root, commit string) error {
	return exec.Command("git", "-C", root, "merge-base", "--is-ancestor", commit, "HEAD").Run()
}

func shortHash(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 12 {
		s = s[:12]
	}
	return s
}

func short(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 160 {
		s = s[:160] + "..."
	}
	return s
}
