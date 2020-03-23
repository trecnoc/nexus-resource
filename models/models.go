package models

import "strings"

type Source struct {
	URL        string `json:"url"`
	Repository string `json:"repository"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Group      string `json:"directory"`
	Regexp     string `json:"regexp"`
}

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

type Version struct {
	Path string `json:"path,omitempty"`
}

type MetadataPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type RespositoryItems struct {
	Items []RepositoryItem `json:"items"`
}

type RepositoryItem struct {
	ID     string                `json:"id"`
	Group  string                `json:"group"`
	Name   string                `json:"name"`
	Assets []RepositoryItemAsset `json:"assets"`
}

type RepositoryItemAsset struct {
	DownloadURL string                       `json:"downloadUrl"`
	ID          string                       `json:"id"`
	Checksum    RepositoryItemAssetsChecksum `json:"checksum"`
}

type RepositoryItemAssetsChecksum struct {
	Sha1   string `json:"sha1"`
	Sha256 string `json:"sha256"`
	Sha512 string `json:"sha512"`
	Md5    string `json:"md5"`
}
