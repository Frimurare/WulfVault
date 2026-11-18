// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package database

import "testing"

// The strings are stored in Files.Size and shown in e-mails, so the format
// must not drift between releases.
func TestFormatFileSize(t *testing.T) {
	cases := map[int64]string{
		-5:                            "-5 B",
		0:                             "0 B",
		1023:                          "1023 B",
		1024:                          "1.0 KB",
		1536:                          "1.5 KB",
		1024*1024 - 1:                 "1024.0 KB",
		1024 * 1024:                   "1.0 MB",
		5*1024*1024*1024 + 1:          "5.0 GB",
		3 * 1024 * 1024 * 1024 * 1024: "3.0 TB",
		1 << 62:                       "4.0 EB",
	}
	for in, want := range cases {
		if got := FormatFileSize(in); got != want {
			t.Errorf("FormatFileSize(%d) = %q, want %q", in, got, want)
		}
	}
}
