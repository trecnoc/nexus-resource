package models

import "strings"

// Source Struct for the Nexus Resource
type Source struct {
	URL        string `json:"url"`
	Repository string `json:"repository"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Group      string `json:"directory"`
	Regexp     string `json:"regexp"`
}

// IsValid validates the provided Source
func (source Source) IsValid() (bool, string) {
	if source.URL == "" {
		return false, "url must be specified"
	}

	if source.Repository == "" {
		return false, "repository must be specified"
	}

	if source.Username == "" {
		return false, "username must be specified"
	}

	if source.Password == "" {
		return false, "password must be specified"
	}

	if source.Group != "" && !strings.HasPrefix(source.Group, "/") {
		return false, "group must start with '/'"
	}

	if source.Regexp != "" && strings.HasPrefix(source.Regexp, "/") {
		return false, "regexp should not start with '/'"
	}

	return true, ""
}

// Version struct
type Version struct {
	Path string `json:"path,omitempty"`
}

// MetadataPair struct
type MetadataPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// RespositoryItems struct is a collection of RepositoryItems
type RespositoryItems struct {
	Items []RepositoryItem `json:"items"`
}

// RepositoryItem struct represent a Component in Nexus
type RepositoryItem struct {
	ID     string                `json:"id"`
	Group  string                `json:"group"`
	Name   string                `json:"name"`
	Assets []RepositoryItemAsset `json:"assets"`
}

// RepositoryItemAsset struct represent an Asset in Nexus
type RepositoryItemAsset struct {
	DownloadURL string                       `json:"downloadUrl"`
	ID          string                       `json:"id"`
	Checksum    RepositoryItemAssetsChecksum `json:"checksum"`
}

// RepositoryItemAssetsChecksum struct represent an Assets Checksum in Nexus
type RepositoryItemAssetsChecksum struct {
	Sha1   string `json:"sha1"`
	Sha256 string `json:"sha256"`
	Sha512 string `json:"sha512"`
	Md5    string `json:"md5"`
}
