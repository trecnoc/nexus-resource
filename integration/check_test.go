package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"github.com/trecnoc/nexus-resource/check"
	"github.com/trecnoc/nexus-resource/models"
)

var _ = Describe("check", func() {
	var (
		command *exec.Cmd
		stdin   *bytes.Buffer
		session *gexec.Session

		expectedExitStatus int
	)

	BeforeEach(func() {
		stdin = &bytes.Buffer{}
		expectedExitStatus = 0

		command = exec.Command(checkPath)
		command.Stdin = stdin
	})

	JustBeforeEach(func() {
		var err error
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		<-session.Exited
		Ω(session.ExitCode()).Should(Equal(expectedExitStatus))
	})

	Context("when we do not provide a previous version", func() {
		var group string
		var checkRequest check.Request

		BeforeEach(func() {
			checkRequest = check.Request{
				Source: models.Source{
					URL:        url,
					Repository: repository,
					Username:   username,
					Password:   password,
				},
				Version: models.Version{},
			}
		})

		Context("with files in the group that do not match", func() {
			Context("with group with no glob", func() {
				BeforeEach(func() {
					group = "/files-in-group-that-do-not-match"

					checkRequest.Source.Group = group
					checkRequest.Source.Regexp = "files-in-group-that-do-not-match/file-does-match-(.*)"

					err := json.NewEncoder(stdin).Encode(checkRequest)
					Ω(err).ShouldNot(HaveOccurred())

					tempFile, err := ioutil.TempFile("", "file-to-upload")
					Ω(err).ShouldNot(HaveOccurred())
					tempFile.Close()

					err = nexusclient.UploadFile(repository, group, "file-does-not-match-1", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = os.Remove(tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					// Let Nexus time to index for search
					time.Sleep(1 * time.Second)
				})

				AfterEach(func() {
					err := nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(group, "/"), "file-does-not-match-1"))
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("returns an empty check response", func() {
					reader := bytes.NewBuffer(session.Out.Contents())

					var response check.Response
					err := json.NewDecoder(reader).Decode(&response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(BeEmpty())
				})
			})

			Context("with group with glob", func() {
				BeforeEach(func() {
					group = "/files-in-group-that-do-not-match/v*/sub"

					checkRequest.Source.Group = group
					checkRequest.Source.Regexp = "files-in-group-that-do-not-match/v(.*)/sub/file-does-match"

					err := json.NewEncoder(stdin).Encode(checkRequest)
					Ω(err).ShouldNot(HaveOccurred())

					tempFile, err := ioutil.TempFile("", "file-to-upload")
					Ω(err).ShouldNot(HaveOccurred())
					tempFile.Close()

					err = nexusclient.UploadFile(repository, "/files-in-group-that-do-not-match/v1/sub", "file-does-not-match", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.UploadFile(repository, "/files-in-group-that-do-not-match/v2/sub", "file-does-not-match", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = os.Remove(tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					// Let Nexus time to index for search
					time.Sleep(1 * time.Second)
				})

				AfterEach(func() {
					err := nexusclient.DeleteFile(repository, "files-in-group-that-do-not-match/v1/sub/file-does-not-match")
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.DeleteFile(repository, "files-in-group-that-do-not-match/v2/sub/file-does-not-match")
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("returns an empty check response", func() {
					reader := bytes.NewBuffer(session.Out.Contents())

					var response check.Response
					err := json.NewDecoder(reader).Decode(&response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(BeEmpty())
				})
			})
		})

		Context("with files in the group that match", func() {
			Context("with group with no glob", func() {
				BeforeEach(func() {
					group = "/files-in-group-that-match"
					checkRequest.Source.Group = group
					checkRequest.Source.Regexp = "files-in-group-that-match/file-does-match-(.*)"

					err := json.NewEncoder(stdin).Encode(checkRequest)
					Ω(err).ShouldNot(HaveOccurred())

					tempFile, err := ioutil.TempFile("", "file-to-upload")
					Ω(err).ShouldNot(HaveOccurred())
					tempFile.Close()

					err = nexusclient.UploadFile(repository, group, "file-does-match-1", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.UploadFile(repository, group, "file-does-match-2", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = os.Remove(tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					// Let Nexus time to index for search
					time.Sleep(1 * time.Second)
				})

				AfterEach(func() {
					err := nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(group, "/"), "file-does-match-1"))
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(group, "/"), "file-does-match-2"))
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("output the path of the latest artifact from the group", func() {
					reader := bytes.NewBuffer(session.Out.Contents())

					var response check.Response
					err := json.NewDecoder(reader).Decode(&response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(Equal(check.Response{
						{
							Path: filepath.Join(strings.TrimPrefix(group, "/"), "/file-does-match-2"),
						},
					}))
				})
			})

			Context("with group with glob", func() {
				Context("with version in group", func() {
					BeforeEach(func() {
						group = "/files-in-group-that-match/v*"
						checkRequest.Source.Group = group
						checkRequest.Source.Regexp = "files-in-group-that-match/v(.*)/file-does-match"

						err := json.NewEncoder(stdin).Encode(checkRequest)
						Ω(err).ShouldNot(HaveOccurred())

						tempFile, err := ioutil.TempFile("", "file-to-upload")
						Ω(err).ShouldNot(HaveOccurred())
						tempFile.Close()

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match/v1", "file-does-match", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match/v2", "file-does-match", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = os.Remove(tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						// Let Nexus time to index for search
						time.Sleep(1 * time.Second)
					})

					AfterEach(func() {
						err := nexusclient.DeleteFile(repository, "files-in-group-that-match/v1/file-does-match")
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.DeleteFile(repository, "files-in-group-that-match/v2/file-does-match")
						Ω(err).ShouldNot(HaveOccurred())
					})

					It("output the path of the latest artifact from the group", func() {
						reader := bytes.NewBuffer(session.Out.Contents())

						var response check.Response
						err := json.NewDecoder(reader).Decode(&response)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(Equal(check.Response{
							{
								Path: "files-in-group-that-match/v2/file-does-match",
							},
						}))
					})
				})
				Context("with version in filename", func() {
					BeforeEach(func() {
						group = "/files-in-group-that-match/*"
						checkRequest.Source.Group = group
						checkRequest.Source.Regexp = "files-in-group-that-match/.*/file-does-match-(.*)"

						err := json.NewEncoder(stdin).Encode(checkRequest)
						Ω(err).ShouldNot(HaveOccurred())

						tempFile, err := ioutil.TempFile("", "file-to-upload")
						Ω(err).ShouldNot(HaveOccurred())
						tempFile.Close()

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match/3", "file-does-match-1", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match/4", "file-does-match-2", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = os.Remove(tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						// Let Nexus time to index for search
						time.Sleep(1 * time.Second)
					})

					AfterEach(func() {
						err := nexusclient.DeleteFile(repository, "files-in-group-that-match/3/file-does-match-1")
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.DeleteFile(repository, "files-in-group-that-match/4/file-does-match-2")
						Ω(err).ShouldNot(HaveOccurred())
					})

					It("output the path of the latest artifact from the group", func() {
						reader := bytes.NewBuffer(session.Out.Contents())

						var response check.Response
						err := json.NewDecoder(reader).Decode(&response)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(Equal(check.Response{
							{
								Path: "files-in-group-that-match/4/file-does-match-2",
							},
						}))
					})
				})
			})

			Context("with group with will use the Continuation Token", func() {
				var itemCount int
				BeforeEach(func() {
					itemCount = 50
					group = "/files-in-a-large-group-that-match"
					checkRequest.Source.Group = group
					checkRequest.Source.Regexp = "files-in-a-large-group-that-match/file-does-match-(.*)"

					err := json.NewEncoder(stdin).Encode(checkRequest)
					Ω(err).ShouldNot(HaveOccurred())

					tempFile, err := ioutil.TempFile("", "file-to-upload")
					Ω(err).ShouldNot(HaveOccurred())
					tempFile.Close()

					for i := 1; i <= itemCount; i++ {
						err = nexusclient.UploadFile(repository, group, fmt.Sprintf("file-does-match-%d", i), tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())
					}

					err = os.Remove(tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					// Let Nexus time to index for search
					time.Sleep(2 * time.Second)
				})

				AfterEach(func() {
					for i := 1; i <= itemCount; i++ {
						err := nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(group, "/"), fmt.Sprintf("file-does-match-%d", i)))
						Ω(err).ShouldNot(HaveOccurred())
					}
				})

				It("output the path of the latest artifact from the group", func() {
					reader := bytes.NewBuffer(session.Out.Contents())

					var response check.Response
					err := json.NewDecoder(reader).Decode(&response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(Equal(check.Response{
						{
							Path: filepath.Join(strings.TrimPrefix(group, "/"), "/file-does-match-"+strconv.Itoa(itemCount)),
						},
					}))
				})
			})

		})
	})

	Context("when we provide a previous version", func() {
		var group string
		var checkRequest check.Request

		BeforeEach(func() {
			checkRequest = check.Request{
				Source: models.Source{
					URL:        url,
					Repository: repository,
					Username:   username,
					Password:   password,
				},
			}
		})

		Context("with files in the group that do not match", func() {
			Context("with group with no glob", func() {
				BeforeEach(func() {
					group = "/files-in-group-that-do-not-match-with-version"

					checkRequest.Source.Group = group
					checkRequest.Source.Regexp = "files-in-group-that-do-not-match-with-version/file-does-match-(.*)"
					checkRequest.Version.Path = "files-in-group-that-do-not-match-with-version/file-does-match-1"

					err := json.NewEncoder(stdin).Encode(checkRequest)
					Ω(err).ShouldNot(HaveOccurred())

					tempFile, err := ioutil.TempFile("", "file-to-upload")
					Ω(err).ShouldNot(HaveOccurred())
					tempFile.Close()

					err = nexusclient.UploadFile(repository, group, "file-does-not-match-1", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = os.Remove(tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					// Let Nexus time to index for search
					time.Sleep(1 * time.Second)
				})

				AfterEach(func() {
					err := nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(group, "/"), "file-does-not-match-1"))
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("returns an empty check response", func() {
					reader := bytes.NewBuffer(session.Out.Contents())

					var response check.Response
					err := json.NewDecoder(reader).Decode(&response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(BeEmpty())
				})
			})

			Context("with group with glob", func() {
				BeforeEach(func() {
					group = "/files-in-group-that-do-not-match-with-version/v*/sub"

					checkRequest.Source.Group = group
					checkRequest.Source.Regexp = "files-in-group-that-do-not-match-with-version/v(.*)/sub/file-does-match"
					checkRequest.Version.Path = "files-in-group-that-do-not-match-with-version/v1/sub/file-does-match"

					err := json.NewEncoder(stdin).Encode(checkRequest)
					Ω(err).ShouldNot(HaveOccurred())

					tempFile, err := ioutil.TempFile("", "file-to-upload")
					Ω(err).ShouldNot(HaveOccurred())
					tempFile.Close()

					err = nexusclient.UploadFile(repository, "/files-in-group-that-do-not-match-with-version/v1/sub", "file-does-not-match", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.UploadFile(repository, "/files-in-group-that-do-not-match-with-version/v2/sub", "file-does-not-match", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = os.Remove(tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					// Let Nexus time to index for search
					time.Sleep(1 * time.Second)
				})

				AfterEach(func() {
					err := nexusclient.DeleteFile(repository, "files-in-group-that-do-not-match-with-version/v1/sub/file-does-not-match")
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.DeleteFile(repository, "files-in-group-that-do-not-match-with-version/v2/sub/file-does-not-match")
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("returns an empty check response", func() {
					reader := bytes.NewBuffer(session.Out.Contents())

					var response check.Response
					err := json.NewDecoder(reader).Decode(&response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(BeEmpty())
				})
			})
		})

		Context("with files in the group that match", func() {
			Context("with group with no glob", func() {
				BeforeEach(func() {
					group = "/files-in-group-that-match-with-version"
					checkRequest.Source.Group = group
					checkRequest.Source.Regexp = "files-in-group-that-match-with-version/file-does-match-(.*)"
					checkRequest.Version.Path = "files-in-group-that-match-with-version/file-does-match-2"

					err := json.NewEncoder(stdin).Encode(checkRequest)
					Ω(err).ShouldNot(HaveOccurred())

					tempFile, err := ioutil.TempFile("", "file-to-upload")
					Ω(err).ShouldNot(HaveOccurred())
					tempFile.Close()

					err = nexusclient.UploadFile(repository, group, "file-does-match-2", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.UploadFile(repository, group, "file-does-match-1", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.UploadFile(repository, group, "file-does-match-3", tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					err = os.Remove(tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					// Let Nexus time to index for search
					time.Sleep(1 * time.Second)
				})

				AfterEach(func() {
					err := nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(group, "/"), "file-does-match-1"))
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(group, "/"), "file-does-match-2"))
					Ω(err).ShouldNot(HaveOccurred())

					err = nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(group, "/"), "file-does-match-3"))
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("output the path of the latest artifact from the group", func() {
					reader := bytes.NewBuffer(session.Out.Contents())

					var response check.Response
					err := json.NewDecoder(reader).Decode(&response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(Equal(check.Response{
						{
							Path: filepath.Join(strings.TrimPrefix(group, "/"), "/file-does-match-2"),
						},
						{
							Path: filepath.Join(strings.TrimPrefix(group, "/"), "/file-does-match-3"),
						},
					}))
				})
			})

			Context("with group with glob", func() {
				Context("with version in group", func() {
					BeforeEach(func() {
						group = "/files-in-group-that-match-with-version/v*"
						checkRequest.Source.Group = group
						checkRequest.Source.Regexp = "files-in-group-that-match-with-version/v(.*)/file-does-match"
						checkRequest.Version.Path = "files-in-group-that-match-with-version/v2/file-does-match"

						err := json.NewEncoder(stdin).Encode(checkRequest)
						Ω(err).ShouldNot(HaveOccurred())

						tempFile, err := ioutil.TempFile("", "file-to-upload")
						Ω(err).ShouldNot(HaveOccurred())
						tempFile.Close()

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match-with-version/v2", "file-does-match", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match-with-version/v1", "file-does-match", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match-with-version/v3", "file-does-match", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = os.Remove(tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						// Let Nexus time to index for search
						time.Sleep(1 * time.Second)
					})

					AfterEach(func() {
						err := nexusclient.DeleteFile(repository, "files-in-group-that-match-with-version/v1/file-does-match")
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.DeleteFile(repository, "files-in-group-that-match-with-version/v2/file-does-match")
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.DeleteFile(repository, "files-in-group-that-match-with-version/v3/file-does-match")
						Ω(err).ShouldNot(HaveOccurred())
					})

					It("output the path of the latest artifact from the group", func() {
						reader := bytes.NewBuffer(session.Out.Contents())

						var response check.Response
						err := json.NewDecoder(reader).Decode(&response)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(Equal(check.Response{
							{
								Path: "files-in-group-that-match-with-version/v2/file-does-match",
							},
							{
								Path: "files-in-group-that-match-with-version/v3/file-does-match",
							},
						}))
					})
				})
				Context("with version in filename", func() {
					BeforeEach(func() {
						group = "/files-in-group-that-match-with-version/*"
						checkRequest.Source.Group = group
						checkRequest.Source.Regexp = "files-in-group-that-match-with-version/.*/file-does-match-(.*)"
						checkRequest.Version.Path = "files-in-group-that-match-with-version/4/file-does-match-2"

						err := json.NewEncoder(stdin).Encode(checkRequest)
						Ω(err).ShouldNot(HaveOccurred())

						tempFile, err := ioutil.TempFile("", "file-to-upload")
						Ω(err).ShouldNot(HaveOccurred())
						tempFile.Close()

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match-with-version/3", "file-does-match-1", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match-with-version/4", "file-does-match-2", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.UploadFile(repository, "/files-in-group-that-match-with-version/5", "file-does-match-3", tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						err = os.Remove(tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())

						// Let Nexus time to index for search
						time.Sleep(1 * time.Second)
					})

					AfterEach(func() {
						err := nexusclient.DeleteFile(repository, "files-in-group-that-match-with-version/3/file-does-match-1")
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.DeleteFile(repository, "files-in-group-that-match-with-version/4/file-does-match-2")
						Ω(err).ShouldNot(HaveOccurred())

						err = nexusclient.DeleteFile(repository, "files-in-group-that-match-with-version/5/file-does-match-3")
						Ω(err).ShouldNot(HaveOccurred())
					})

					It("output the path of the latest artifact from the group", func() {
						reader := bytes.NewBuffer(session.Out.Contents())

						var response check.Response
						err := json.NewDecoder(reader).Decode(&response)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(Equal(check.Response{
							{
								Path: "files-in-group-that-match-with-version/4/file-does-match-2",
							},
							{
								Path: "files-in-group-that-match-with-version/5/file-does-match-3",
							},
						}))
					})
				})
			})

			Context("with group with will use the Continuation Token", func() {
				var itemCount int
				BeforeEach(func() {
					itemCount = 50
					group = "/files-in-a-large-group-that-match-with-version"
					checkRequest.Source.Group = group
					checkRequest.Source.Regexp = "files-in-a-large-group-that-match-with-version/file-does-match-(.*)"
					checkRequest.Version.Path = "files-in-a-large-group-that-match-with-version/file-does-match-48"

					err := json.NewEncoder(stdin).Encode(checkRequest)
					Ω(err).ShouldNot(HaveOccurred())

					tempFile, err := ioutil.TempFile("", "file-to-upload")
					Ω(err).ShouldNot(HaveOccurred())
					tempFile.Close()

					for i := 1; i <= itemCount; i++ {
						err = nexusclient.UploadFile(repository, group, fmt.Sprintf("file-does-match-%d", i), tempFile.Name())
						Ω(err).ShouldNot(HaveOccurred())
					}

					err = os.Remove(tempFile.Name())
					Ω(err).ShouldNot(HaveOccurred())

					// Let Nexus time to index for search
					time.Sleep(2 * time.Second)
				})

				AfterEach(func() {
					for i := 1; i <= itemCount; i++ {
						err := nexusclient.DeleteFile(repository, filepath.Join(strings.TrimPrefix(group, "/"), fmt.Sprintf("file-does-match-%d", i)))
						Ω(err).ShouldNot(HaveOccurred())
					}
				})

				It("output the path of the latest artifact from the group", func() {
					reader := bytes.NewBuffer(session.Out.Contents())

					var response check.Response
					err := json.NewDecoder(reader).Decode(&response)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(Equal(check.Response{
						{
							Path: filepath.Join(strings.TrimPrefix(group, "/"), "/file-does-match-48"),
						},
						{
							Path: filepath.Join(strings.TrimPrefix(group, "/"), "/file-does-match-49"),
						},
						{
							Path: filepath.Join(strings.TrimPrefix(group, "/"), "/file-does-match-50"),
						},
					}))
				})
			})

		})
	})

})
