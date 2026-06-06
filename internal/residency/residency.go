package residency

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/agentic-paywall/agentic-paywall/internal/config"
)

type Statement struct {
	Region              string   `json:"region"`
	CollectedFields     []string `json:"collected_fields"`
	StorageLocation     string   `json:"storage_location"`
	SubProcessors       []string `json:"sub_processors"`
	RawIPRetained       bool     `json:"raw_ip_retained"`
	CrossBorderTransfer bool     `json:"cross_border_transfer"`
}

func FromConfig(cfg *config.Config) Statement {
	return Statement{
		Region: cfg.DataResidency.Region,
		CollectedFields: []string{
			"ts",
			"type",
			"payment_id",
			"resource_path",
			"amount",
			"currency",
			"agent_id",
			"decision",
		},
		StorageLocation:     cfg.DataResidency.StorageLocation,
		SubProcessors:       append([]string(nil), cfg.DataResidency.AllowedSubProcessors...),
		RawIPRetained:       false,
		CrossBorderTransfer: false,
	}
}

func WriteMarkdown(path string, statement Statement) error {
	var b strings.Builder
	b.WriteString("# Live Data Residency Statement\n\n")
	b.WriteString("This file is generated from the running deployment configuration.\n\n")
	b.WriteString(fmt.Sprintf("- Region: %s\n", statement.Region))
	b.WriteString(fmt.Sprintf("- Storage location: %s\n", statement.StorageLocation))
	b.WriteString(fmt.Sprintf("- Raw IP retained: %t\n", statement.RawIPRetained))
	b.WriteString(fmt.Sprintf("- Cross-border transfer: %t\n", statement.CrossBorderTransfer))
	b.WriteString("- Collected fields:\n")
	for _, field := range statement.CollectedFields {
		b.WriteString(fmt.Sprintf("  - %s\n", field))
	}
	b.WriteString("- Sub-processors:\n")
	for _, processor := range statement.SubProcessors {
		b.WriteString(fmt.Sprintf("  - %s\n", processor))
	}
	b.WriteString("\n## Machine-readable statement\n\n")
	data, err := json.MarshalIndent(statement, "", "  ")
	if err != nil {
		return err
	}
	b.WriteString("```json\n")
	b.Write(data)
	b.WriteString("\n```\n")
	return os.WriteFile(path, []byte(b.String()), 0o600)
}
