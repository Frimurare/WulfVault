// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU General Public License v3.0 (GPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

import (
	"github.com/forceu/gokapi/internal/test"
	"testing"
)

func TestIsAwsProvided(t *testing.T) {
	config := AwsConfig{}
	test.IsEqualBool(t, config.IsAllProvided(), false)
	config = AwsConfig{
		Bucket:    "test",
		Region:    "test",
		Endpoint:  "",
		KeyId:     "test",
		KeySecret: "test",
	}
	test.IsEqualBool(t, config.IsAllProvided(), true)
}
