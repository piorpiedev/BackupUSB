package archive

import (
	"fmt"
)

func FormatByteCount(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%-6dB", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	v := float64(b) / float64(div)
	d := fmt.Sprintf("%.2f", v)[:4]
	if d[3] == '.' {
		d = d[:3] + " "
	}

	return fmt.Sprintf("%s %cB", d, "KMGTPE"[exp])
}
