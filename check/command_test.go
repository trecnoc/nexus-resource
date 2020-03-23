package check_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/trecnoc/nexus-resource/check"

	"github.com/trecnoc/nexus-resource/fakes"
	"github.com/trecnoc/nexus-resource/models"
)

var _ = Describe("Check Command", func() {
	Describe("running the command", func() {
		var (
			tmpPath string
			request Request

			nexusclient *fakes.FakeNexusClient
			command     *Command
		)

		BeforeEach(func() {
			var err error
			tmpPath, err = ioutil.TempDir("", "check_command")
			Ω(err).ShouldNot(HaveOccurred())

			request = Request{
				Source: models.Source{
					URL:        "http://nexus-url.com",
					Repository: "repository-name",
					Username:   "user",
					Password:   "password",
				},
			}

			nexusclient = &fakes.FakeNexusClient{}
			command = NewCommand(nexusclient)
		})

		AfterEach(func() {
			err := os.RemoveAll(tmpPath)
			Ω(err).ShouldNot(HaveOccurred())
		})

		Context("when there is no previous version", func() {
			Context("when a non-glob group is used", func() {
				It("includes the latest version only", func() {
					request.Version.Path = ""
					request.Source.Group = "/files"
					request.Source.Regexp = "files/abc-(.*).tgz"

					nexusclient.ListFilesReturns([]string{
						"files/abc-0.0.1.tgz",
						"files/abc-2.33.333.tgz",
						"files/abc-2.4.3.tgz",
						"files/abc-3.53.tgz",
					}, nil)

					response, err := command.Run(request)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(HaveLen(1))
					Ω(response).Should(ConsistOf(
						models.Version{
							Path: "files/abc-3.53.tgz",
						},
					))
				})
			})
			Context("when a globbed group is used", func() {
				It("includes the latest version only", func() {
					request.Version.Path = ""
					request.Source.Group = "/files/*/sub"
					request.Source.Regexp = "files/v(.*)/sub/abc.tgz"

					nexusclient.ListFilesReturns([]string{
						"files/v0.0.1/sub/abc.tgz",
						"files/v2.33.333/sub/abc.tgz",
						"files/v2.4.3/sub/abc.tgz",
						"files/v3.53/sub/abc.tgz",
					}, nil)

					response, err := command.Run(request)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(response).Should(HaveLen(1))
					Ω(response).Should(ConsistOf(
						models.Version{
							Path: "files/v3.53/sub/abc.tgz",
						},
					))
				})
			})

			Context("when the regexp does not match anything", func() {
				Context("when a non-glob group is used", func() {
					It("does not explode", func() {
						request.Source.Group = "/files"
						request.Source.Regexp = "files/missing-(.*).tgz"

						nexusclient.ListFilesReturns([]string{
							"files/abc-0.0.1.tgz",
							"files/abc-2.33.333.tgz",
							"files/abc-2.4.3.tgz",
							"files/abc-3.53.tgz",
						}, nil)

						response, err := command.Run(request)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(HaveLen(0))
					})
				})

				Context("when a globbed group is used", func() {
					It("does not explode", func() {
						request.Source.Group = "/files/*/sub"
						request.Source.Regexp = "files/v(.*)/sub/missing-(.*).tgz"

						nexusclient.ListFilesReturns([]string{
							"files/v0.0.1/sub/abc.tgz",
							"files/v2.33.333/sub/abc.tgz",
							"files/v2.4.3/sub/abc.tgz",
							"files/v3.53/sub/abc.tgz",
						}, nil)

						response, err := command.Run(request)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(HaveLen(0))
					})
				})
			})

			Context("when the regex does not match the previous version", func() {
				Context("when a non-glob group is used", func() {
					It("returns the latest version that matches the regex", func() {
						request.Version.Path = "files/abc-0.0.1.tgz"
						request.Source.Group = "/files"
						request.Source.Regexp = `files/abc-(2\.33.*).tgz`

						nexusclient.ListFilesReturns([]string{
							"files/abc-0.0.1.tgz",
							"files/abc-2.33.333.tgz",
							"files/abc-2.4.3.tgz",
							"files/abc-3.53.tgz",
						}, nil)

						response, err := command.Run(request)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(HaveLen(1))
						Expect(response).To(ConsistOf(models.Version{Path: "files/abc-2.33.333.tgz"}))
					})
				})

				Context("when a globbed group is used", func() {
					It("returns the latest version that matches the regex", func() {
						request.Version.Path = "files/v0.0.1/sub/abc.tgz"
						request.Source.Group = "/files/*/sub"
						request.Source.Regexp = `files/v(2\.33.*)/sub/abc.tgz`

						nexusclient.ListFilesReturns([]string{
							"files/v0.0.1/sub/abc.tgz",
							"files/v2.33.333/sub/abc.tgz",
							"files/v2.4.3/sub/abc.tgz",
							"files/v3.53/sub/abc.tgz",
						}, nil)

						response, err := command.Run(request)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(HaveLen(1))
						Expect(response).To(ConsistOf(models.Version{Path: "files/v2.33.333/sub/abc.tgz"}))
					})
				})
			})
		})

		Context("when there is a previous version", func() {
			Context("when using regex that matches the provided version", func() {
				Context("when a non-glob group is used", func() {
					It("includes all versions from the previous one and the current one", func() {
						request.Version.Path = "files/abc-2.4.3.tgz"
						request.Source.Group = "/files"
						request.Source.Regexp = "files/abc-(.*).tgz"

						nexusclient.ListFilesReturns([]string{
							"files/abc-0.0.1.tgz",
							"files/abc-2.33.333.tgz",
							"files/abc-2.4.3.tgz",
							"files/abc-3.53.tgz",
						}, nil)

						response, err := command.Run(request)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(HaveLen(3))
						Ω(response).Should(ConsistOf(
							models.Version{
								Path: "files/abc-2.4.3.tgz",
							},
							models.Version{
								Path: "files/abc-2.33.333.tgz",
							},
							models.Version{
								Path: "files/abc-3.53.tgz",
							},
						))
					})
				})
				Context("when a globbed group is used", func() {
					It("includes all versions from the previous one and the current one", func() {
						request.Version.Path = "files/v2.4.3/sub/abc.tgz"
						request.Source.Group = "/files/v*/sub"
						request.Source.Regexp = "files/v(.*)/sub/abc.tgz"

						nexusclient.ListFilesReturns([]string{
							"files/v0.0.1/sub/abc.tgz",
							"files/v2.33.333/sub/abc.tgz",
							"files/v2.4.3/sub/abc.tgz",
							"files/v3.53/sub/abc.tgz",
						}, nil)

						response, err := command.Run(request)
						Ω(err).ShouldNot(HaveOccurred())

						Ω(response).Should(HaveLen(3))
						Ω(response).Should(ConsistOf(
							models.Version{
								Path: "files/v2.4.3/sub/abc.tgz",
							},
							models.Version{
								Path: "files/v2.33.333/sub/abc.tgz",
							},
							models.Version{
								Path: "files/v3.53/sub/abc.tgz",
							},
						))
					})
				})
			})
		})
	})
})
