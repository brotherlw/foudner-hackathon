package approval

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Prompter struct {
	in  io.Reader
	out io.Writer
}

func NewPrompter() *Prompter {
	return &Prompter{
		in:  os.Stdin,
		out: os.Stderr,
	}
}

func (p *Prompter) RequireApproval(amount, threshold float64, purpose string) error {
	if amount <= threshold {
		return nil
	}
	fmt.Fprintf(p.out, "Payment approval required: %.2f EUR for %q (threshold %.2f EUR). Proceed? [y/N]: ", amount, purpose, threshold)
	reader := bufio.NewReader(p.in)
	line, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("approval prompt failed: %w", err)
	}
	answer := strings.TrimSpace(strings.ToLower(line))
	if answer != "y" && answer != "yes" {
		return fmt.Errorf("payment rejected by user")
	}
	return nil
}
