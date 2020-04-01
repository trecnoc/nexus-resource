package versions

import (
	"regexp"
	"sort"

	"github.com/cppforlife/go-semi-semantic/version"
	"github.com/trecnoc/nexus-resource"
	"github.com/trecnoc/nexus-resource/models"
	"github.com/trecnoc/nexus-resource/utils"
)

// Match paths against a provided pattern by anchoring it
func Match(paths []string, pattern string) ([]string, error) {
	return MatchUnanchored(paths, "^"+pattern+"$")
}

// MatchUnanchored paths against a pattern
func MatchUnanchored(paths []string, pattern string) ([]string, error) {
	matched := []string{}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return matched, err
	}

	for _, path := range paths {
		match := regex.MatchString(path)

		if match {
			matched = append(matched, path)
		}
	}

	return matched, nil
}

// Extract an version from a path with a provided pattern
func Extract(path string, pattern string) (Extraction, bool) {
	compiled := regexp.MustCompile(pattern)
	matches := compiled.FindStringSubmatch(path)

	var match string
	if len(matches) < 2 { // whole string and match
		return Extraction{}, false
	} else if len(matches) == 2 {
		match = matches[1]
	} else if len(matches) > 2 { // many matches
		names := compiled.SubexpNames()
		index := sliceIndex(names, "version")

		if index > 0 {
			match = matches[index]
		} else {
			match = matches[1]
		}
	}

	ver, err := version.NewVersionFromString(match)
	if err != nil {
		panic("version number was not valid: " + err.Error())
	}

	extraction := Extraction{
		Path:          path,
		Version:       ver,
		VersionNumber: match,
	}

	return extraction, true
}

func sliceIndex(haystack []string, needle string) int {
	for i, element := range haystack {
		if element == needle {
			return i
		}
	}

	return -1
}

// Extractions type
type Extractions []Extraction

func (e Extractions) Len() int {
	return len(e)
}

func (e Extractions) Less(i int, j int) bool {
	return e[i].Version.IsLt(e[j].Version)
}

func (e Extractions) Swap(i int, j int) {
	e[i], e[j] = e[j], e[i]
}

// Extraction struct for a path/version combination
type Extraction struct {
	// path to nexus artifact in repository
	Path string

	// parsed version
	Version version.Version

	// the raw version match
	VersionNumber string
}

// GetRepositoryItemVersions returns the Extractions for a provided Source
func GetRepositoryItemVersions(client nexusresource.NexusClient, source models.Source) Extractions {
	paths, err := client.ListFiles(source.Repository, source.Group)
	if err != nil {
		utils.Fatal("listing files", err)
	}

	matchingPaths, err := Match(paths, source.Regexp)
	if err != nil {
		utils.Fatal("finding matches", err)
	}

	var extractions = make(Extractions, 0, len(matchingPaths))
	for _, path := range matchingPaths {
		extraction, ok := Extract(path, source.Regexp)

		if ok {
			extractions = append(extractions, extraction)
		}
	}

	sort.Sort(extractions)

	return extractions
}
