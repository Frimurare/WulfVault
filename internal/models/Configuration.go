// Sharecare - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

import (
	"encoding/json"
	"log"
)

// Configuration is a struct that contains the global configuration
type Configuration struct {
	Authentication      AuthenticationConfig `json:"Authentication"`
	Branding            Branding             `json:"Branding"`
	Port                string               `json:"Port"`
	ServerUrl           string               `json:"ServerUrl"`
	RedirectUrl         string               `json:"RedirectUrl"`
	PublicName          string               `json:"PublicName"`
	DataDir             string               `json:"DataDir"`
	DatabaseUrl         string               `json:"DatabaseUrl"`
	ConfigVersion       int                  `json:"ConfigVersion"`
	MaxFileSizeMB       int                  `json:"MaxFileSizeMB"`
	MaxMemory           int                  `json:"MaxMemory"`
	ChunkSize           int                  `json:"ChunkSize"`
	MaxParallelUploads  int                  `json:"MaxParallelUploads"`
	LengthId            int                  `json:"-"`
	LengthHotlinkId     int                  `json:"-"`
	MinLengthPassword   int                  `json:"-"`
	Encryption          Encryption           `json:"Encryption"`
	UseSsl              bool                 `json:"UseSsl"`
	PicturesAlwaysLocal bool                 `json:"PicturesAlwaysLocal"`
	SaveIp              bool                 `json:"SaveIp"`
	IncludeFilename     bool                 `json:"IncludeFilename"`
	RequireDownloadAuth bool                 `json:"RequireDownloadAuth"` // If true, downloads require authentication by default
}

// Encryption hold information about the encryption used on this file
type Encryption struct {
	Level        int
	Cipher       []byte
	Salt         string
	Checksum     string
	ChecksumSalt string
}

// ToJson returns an idented JSon representation
func (c Configuration) ToJson() []byte {
	result, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Fatal("Error encoding configuration:", err)
	}
	return result
}

// ToString returns the object as an unidented Json string used for test units
func (c Configuration) ToString() string {
	result, err := json.Marshal(c)
	if err != nil {
		log.Fatal(err)
	}
	return string(result)
}
