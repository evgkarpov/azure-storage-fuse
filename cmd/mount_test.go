package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/azure-storage-fuse/v2/common"
	"github.com/Azure/azure-storage-fuse/v2/common/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var configMountTest string = `
logging:
  type: syslog
file_cache:
  path: fileCachePath
libfuse:
  attribute-expiration-sec: 120
  entry-expiration-sec: 60
azstorage:
  account-name: myAccountName
  account-key: myAccountKey
  mode: key
  endpoint: myEndpoint
  container: myContainer
components:
  - libfuse
  - file_cache
  - attr_cache
  - azstorage
`

var confFileMntTest string

type mountTestSuite struct {
	suite.Suite
	assert *assert.Assertions
}

func (suite *mountTestSuite) SetupTest() {
	suite.assert = assert.New(suite.T())
	osExit = suite.testOsExit
	err := log.SetDefaultLogger("silent", common.LogConfig{Level: common.ELogLevel.LOG_DEBUG()})
	if err != nil {
		panic("Unable to set silent logger as default.")
	}
}

func (suite *mountTestSuite) cleanupTest() {
	resetCLIFlags(*mountCmd)
}

func (suite *mountTestSuite) testOsExit(code int) {
	suite.assert.Equal(1, code)
}

// mount failure test where the mount directory does not exists
func (suite *mountTestSuite) TestMountDirNotExists() {
	defer suite.cleanupTest()

	tempDir := randomString(8)
	_, err := executeCommandC(rootCmd, "mount", tempDir, fmt.Sprintf("--config-file=%s", confFileMntTest))
	suite.assert.Nil(err)
}

// mount failure test where the mount directory is not empty
func (suite *mountTestSuite) TestMountDirNotEmpty() {
	defer suite.cleanupTest()

	mntDir, err := ioutil.TempDir("", "mntdir")
	suite.assert.Nil(err)
	tempDir := filepath.Join(mntDir, "tempdir")

	err = os.MkdirAll(tempDir, 0777)
	suite.assert.Nil(err)
	defer os.RemoveAll(mntDir)

	_, err = executeCommandC(rootCmd, "mount", mntDir, fmt.Sprintf("--config-file=%s", confFileMntTest))
	suite.assert.Nil(err)
}

// mount failure test where the mount path is not provided
func (suite *mountTestSuite) TestMountPathNotProvided() {
	defer suite.cleanupTest()

	_, err := executeCommandC(rootCmd, "mount", "", fmt.Sprintf("--config-file=%s", confFileMntTest))
	suite.assert.Nil(err)
}

// mount failure test where the config file type is unsupported
func (suite *mountTestSuite) TestUnsupportedConfigFileType() {
	defer suite.cleanupTest()

	mntDir, err := ioutil.TempDir("", "mntdir")
	suite.assert.Nil(err)
	defer os.RemoveAll(mntDir)

	_, err = executeCommandC(rootCmd, "mount", mntDir, "--config-file=cfgInvalid.yam")
	suite.assert.Nil(err)
}

// mount failure test where the config file is not present
func (suite *mountTestSuite) TestConfigFileNotFound() {
	defer suite.cleanupTest()

	mntDir, err := ioutil.TempDir("", "mntdir")
	suite.assert.Nil(err)
	defer os.RemoveAll(mntDir)

	_, err = executeCommandC(rootCmd, "mount", mntDir, "--config-file=cfgNotFound.yaml")
	suite.assert.Nil(err)
}

// mount failure test where config file is not provided
func (suite *mountTestSuite) TestConfigFileNotProvided() {
	defer suite.cleanupTest()

	mntDir, err := ioutil.TempDir("", "mntdir")
	suite.assert.Nil(err)
	defer os.RemoveAll(mntDir)

	_, err = executeCommandC(rootCmd, "mount", mntDir)
	suite.assert.Nil(err)
}

func TestMountCommand(t *testing.T) {
	defer func() {
		osExit = os.Exit
	}()

	confFile, err := ioutil.TempFile("", "conf*.yaml")
	if err != nil {
		t.Error("Failed to create config file")
	}
	confFileMntTest = confFile.Name()
	defer os.Remove(confFileMntTest)

	_, err = confFile.WriteString(configMountTest)
	if err != nil {
		t.Error("Failed to write to config file")
	}

	suite.Run(t, new(mountTestSuite))
}
