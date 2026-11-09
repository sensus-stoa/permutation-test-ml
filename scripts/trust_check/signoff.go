package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// doSignoff runs the human verdict layer: it refuses to start without a
// green auto-check, walks the signer through every claim, and writes
// trust/signoff.md naming the signer, the date and the exact report
// (by sha256) the signature applies to. A green auto-check without this
// file is explicitly NOT a verdict.
func doSignoff(root, signer, reportPath string, rep *Report) error {
	if strings.TrimSpace(signer) == "" {
		return fmt.Errorf("--signer is required: a sign-off without a name is not a signature")
	}
	repBytes, err := os.ReadFile(reportPath)
	if err != nil {
		return fmt.Errorf("report unavailable: %w", err)
	}
	sum := sha256.Sum256(repBytes)
	repHash := hex.EncodeToString(sum[:])

	fmt.Printf("Human sign-off over %d claims (mechanics already green).\n", len(rep.Claims))
	fmt.Println("Confirm = Space/Enter, skip = s, abort = q")

	var confirmed, skipped []string
	for _, c := range rep.Claims {
		fmt.Printf("[%s | %s] %s  -> ", c.ID, c.Status, claimSummary(c.Verification))
		confirm, skip, abort, err := readAnswer()
		if err != nil {
			return err
		}
		if abort {
			return fmt.Errorf("aborted by signer at claim %s", c.ID)
		}
		if !confirm || skip {
			fmt.Println("skipped")
			skipped = append(skipped, c.ID)
			continue
		}
		fmt.Println("confirmed")
		confirmed = append(confirmed, c.ID)
	}

	var b strings.Builder
	b.WriteString("# Trust Record sign-off\n\n")
	b.WriteString(fmt.Sprintf("- Date: %s\n", time.Now().Format("2006-01-02 15:04 MST")))
	b.WriteString(fmt.Sprintf("- Signer: %s\n", signer))
	b.WriteString(fmt.Sprintf("- Auto-check report: `%s` (sha256 `%s`, overall PASS)\n", filepathBase(reportPath), repHash))
	b.WriteString(fmt.Sprintf("- Confirmed claims: %s\n", orNone(confirmed)))
	b.WriteString(fmt.Sprintf("- Skipped: %s\n", orNone(skipped)))
	b.WriteString("\nRule of order: this signature applies only to the report hashed above;\n")
	b.WriteString("any change to the record or object invalidates it and requires a re-run\n")
	b.WriteString("of the auto-check plus a fresh sign-off. Auto-PASS alone is not a verdict.\n")

	sigPath := root + "/trust/signoff.md"
	if err := os.WriteFile(sigPath, []byte(b.String()), 0o644); err != nil {
		return err
	}
	fmt.Printf("sign-off written: %s\n", sigPath)
	return nil
}

func claimSummary(v string) string {
	v = strings.TrimSpace(v)
	if len(v) > 90 {
		v = v[:90] + "..."
	}
	return v
}

func filepathBase(p string) string {
	i := strings.LastIndexByte(p, '/')
	if i >= 0 {
		return p[i+1:]
	}
	return p
}

func orNone(list []string) string {
	if len(list) == 0 {
		return "none"
	}
	return strings.Join(list, ", ")
}

// readAnswer: single keypress on a terminal (space/enter=confirm, s=skip,
// q=abort); line mode when stdin is piped (empty or y=confirm).
func readAnswer() (confirm, skip, abort bool, err error) {
	fi, _ := os.Stdin.Stat()
	if fi != nil && fi.Mode()&os.ModeCharDevice != 0 {
		raw := exec.Command("stty", "raw", "-echo")
		raw.Stdin = os.Stdin
		_ = raw.Run()
		defer func() {
			sane := exec.Command("stty", "sane")
			sane.Stdin = os.Stdin
			_ = sane.Run()
		}()
		buf := make([]byte, 1)
		for {
			if _, e := os.Stdin.Read(buf); e != nil {
				return false, false, false, e
			}
			switch buf[0] {
			case ' ', '\r', '\n':
				return true, false, false, nil
			case 's', 'S':
				return false, true, false, nil
			case 'q', 'Q', 3: // q / Ctrl-C
				return false, false, true, nil
			default:
				// ignore other keys, keep waiting
			}
		}
	}
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return false, false, false, sc.Err()
	}
	switch strings.TrimSpace(strings.ToLower(sc.Text())) {
	case "", "y", "yes":
		return true, false, false, nil
	case "s", "skip":
		return false, true, false, nil
	case "q", "quit":
		return false, false, true, nil
	default:
		fmt.Println("(empty/Enter = confirm, s = skip, q = abort)")
		return readAnswer()
	}
}
