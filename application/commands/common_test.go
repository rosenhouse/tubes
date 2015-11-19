package commands_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rosenhouse/tubes/application"
	"github.com/rosenhouse/tubes/application/commands"
)

func expectAreSameDirectory(dir1, dir2 string) {
	// we can't just compare equality of strings, because one may be a symlink
	// this happens with temp dir on mac

	filename := fmt.Sprintf("%x", rand.Int31())
	data := []byte(fmt.Sprintf("%x", rand.Int31()))
	Expect(ioutil.WriteFile(filepath.Join(dir1, filename), data, 0600)).To(Succeed())
	actualData, err := ioutil.ReadFile(filepath.Join(dir2, filename))
	Expect(err).NotTo(HaveOccurred())
	Expect(actualData).To(Equal(data))
}

var _ = Describe("InitApp", func() {
	var (
		workingDir string
	)

	BeforeEach(func() {
		var err error
		workingDir, err = ioutil.TempDir("", "tubes-command-unit-test-")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.Chdir(workingDir)).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.Chdir(thisDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	Context("when the state directory is not set", func() {
		It("should create a subdirectory of the working directory", func() {
			options := commands.CLIOptions{
				Name: "some-stack-name",
				AWSConfig: commands.AWSConfig{
					Region:    "some-region",
					AccessKey: "some-access-key",
					SecretKey: "some-secret-key",
				},
			}

			app, err := options.InitApp(nil)
			Expect(err).NotTo(HaveOccurred())

			configStore := app.ConfigStore.(*application.FilesystemConfigStore)
			expectedConfigRootDir := filepath.Join(workingDir, "environments", "some-stack-name")
			expectAreSameDirectory(configStore.RootDir, expectedConfigRootDir)
		})
	})

	Context("when the state directory is set", func() {
		Context("when set to a non-existent directory", func() {
			It("should return an error", func() {
				options := commands.CLIOptions{
					Name: "some-stack-name",
					AWSConfig: commands.AWSConfig{
						Region:    "some-region",
						AccessKey: "some-access-key",
						SecretKey: "some-secret-key",
					},
					StateDir: fmt.Sprintf("-nope-%x-nope", rand.Int31()),
				}

				_, err := options.InitApp(nil)
				Expect(err).To(MatchError(ContainSubstring("state directory not found")))
			})
		})
		Context("when set to a file instead of a directory", func() {
			It("should return an error", func() {
				someFilePath := filepath.Join(workingDir, "this-exists")
				Expect(ioutil.WriteFile(someFilePath, []byte("whatever"), 0600)).To(Succeed())
				options := commands.CLIOptions{
					Name: "some-stack-name",
					AWSConfig: commands.AWSConfig{
						Region:    "some-region",
						AccessKey: "some-access-key",
						SecretKey: "some-secret-key",
					},
					StateDir: someFilePath,
				}

				_, err := options.InitApp(nil)
				Expect(err).To(MatchError(ContainSubstring("state directory not a directory")))
			})
		})
	})
})
