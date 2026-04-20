package vcs

import (
	"runtime/debug"
	"time"
)

func Version() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	var rev string
	var buildTime string
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			// Take the first 7 characters for a short hash
			if len(s.Value) > 7 {
				rev = s.Value[:7]
			} else {
				rev = s.Value
			}
		case "vcs.time":
			// Parse the RFC3339 time and format as YYYYMMDD
			t, err := time.Parse(time.RFC3339, s.Value)
			if err == nil {
				buildTime = t.Format("200601021504")
			}
		}
	}

	if buildTime != "" && rev != "" {
		return buildTime + "-" + rev
	}

	// Fallback to the main version (stripping the leading 'v' if present)
	v := bi.Main.Version
	if len(v) > 0 && v[0] == 'v' {
		return v[1:]
	}

	return v
}
