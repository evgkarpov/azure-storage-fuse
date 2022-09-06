/*
    _____           _____   _____   ____          ______  _____  ------
   |     |  |      |     | |     | |     |     | |       |            |
   |     |  |      |     | |     | |     |     | |       |            |
   | --- |  |      |     | |-----| |---- |     | |-----| |-----  ------
   |     |  |      |     | |     | |     |     |       | |       |
   | ____|  |_____ | ____| | ____| |     |_____|  _____| |_____  |_____


   Licensed under the MIT License <http://opensource.org/licenses/MIT>.

   Copyright © 2020-2022 Microsoft Corporation. All rights reserved.
   Author : <blobfusedev@microsoft.com>

   Permission is hereby granted, free of charge, to any person obtaining a copy
   of this software and associated documentation files (the "Software"), to deal
   in the Software without restriction, including without limitation the rights
   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
   copies of the Software, and to permit persons to whom the Software is
   furnished to do so, subject to the following conditions:

   The above copyright notice and this permission notice shall be included in all
   copies or substantial portions of the Software.

   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
   SOFTWARE
*/

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
