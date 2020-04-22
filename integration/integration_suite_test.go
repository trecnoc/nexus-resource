package integration_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"github.com/trecnoc/nexus-resource"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
}

var url = os.Getenv("NEXUS_TESTING_URL")
var username = os.Getenv("NEXUS_TESTING_USERNAME")
var password = os.Getenv("NEXUS_TESTING_PASSWORD")
var repository = os.Getenv("NEXUS_TESTING_REPOSITORY")
var timeout = os.Getenv("NEXUS_TESTING_TIMEOUT")
var debug = os.Getenv("NEXUS_TESTING_DEBUG")

var nexusclient nexusresource.NexusClient
var checkPath string
var inPath string
var outPath string

type suiteData struct {
	CheckPath string
	InPath    string
	OutPath   string
}

func findOrCreate(binName string) string {
	resourcePath := "/opt/resource/" + binName
	if _, err := os.Stat(resourcePath); err == nil {
		return resourcePath
	} else {
		path, err := gexec.Build("github.com/trecnoc/nexus-resource/cmd/" + binName)
		Ω(err).ShouldNot(HaveOccurred())
		return path
	}
}

var _ = SynchronizedBeforeSuite(func() []byte {
	cp := findOrCreate("check")
	ip := findOrCreate("in")
	op := findOrCreate("out")

	data, err := json.Marshal(suiteData{
		CheckPath: cp,
		InPath:    ip,
		OutPath:   op,
	})

	Ω(err).ShouldNot(HaveOccurred())

	return data
}, func(data []byte) {
	var sd suiteData
	err := json.Unmarshal(data, &sd)
	Ω(err).ShouldNot(HaveOccurred())

	checkPath = sd.CheckPath
	inPath = sd.InPath
	outPath = sd.OutPath

	if url != "" {
		Ω(url).ShouldNot(BeEmpty(), "must specify $NEXUS_TESTING_URL")
		Ω(username).ShouldNot(BeEmpty(), "must specify $NEXUS_TESTING_USERNAME")
		Ω(password).ShouldNot(BeEmpty(), "must specify $NEXUS_TESTING_PASSWORD")
		Ω(repository).ShouldNot(BeEmpty(), "must specify $NEXUS_TESTING_REPOSITORY")

		timeoutInt, err := strconv.Atoi(timeout)
		if err != nil {
			timeoutInt = 0
		}

		debugBool, err := strconv.ParseBool(debug)
		if err != nil {
			debugBool = false
		}

		nexusclient = nexusresource.NewNexusClient(url, username, password, timeoutInt, debugBool)
	}
})

var _ = BeforeEach(func() {
	if nexusclient == nil {
		Skip("Environment variables need to be set for running integration tests")
	}
})

var _ = SynchronizedAfterSuite(func() {}, func() {
	gexec.CleanupBuildArtifacts()
})

func TestIn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}
