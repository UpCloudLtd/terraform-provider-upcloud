package loadbalancer

import (
	"bytes"
	"encoding/base64"
	"encoding/pem"
	"strings"

	"github.com/UpCloudLtd/terraform-provider-upcloud/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// normalizeCertificate takes a base64-encoded PEM certificate (or chain) and returns
// a normalized version with comments, whitespace, and blank lines removed.
// Only the PEM blocks are preserved.
func normalizeCertificate(encoded string) (string, diag.Diagnostics) {
	var respDiagnostics diag.Diagnostics

	if encoded == "" {
		return "", respDiagnostics
	}

	// Strip whitespace from base64 string
	cleaned := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			return -1 // remove
		}
		return r
	}, encoded)

	decoded, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		respDiagnostics.AddError(
			"Unable to decode",
			utils.ErrorDiagnosticDetail(err),
		)
		return "", respDiagnostics
	}

	// Parse all PEM blocks
	var blocks []*pem.Block
	rest := decoded
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		blocks = append(blocks, block)
	}

	if len(blocks) == 0 {
		respDiagnostics.AddError(
			"Unable to parse",
			"No valid PEM blocks found",
		)
		return "", respDiagnostics
	}

	// Re-encode PEM blocks without extra whitespace
	var buf bytes.Buffer
	for _, block := range blocks {
		if err := pem.Encode(&buf, block); err != nil {
			respDiagnostics.AddError(
				"Unable to encode",
				utils.ErrorDiagnosticDetail(err),
			)
			return "", respDiagnostics
		}
	}

	// Encode back to base64
	return base64.StdEncoding.EncodeToString(buf.Bytes()), respDiagnostics
}
