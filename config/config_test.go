package config

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type testConfigStruct struct {
	AuthKey1    string `env-required:"true" yaml:"auth_key" env:"AUTH_KEY"`
	AuthRID     string `env-required:"true" yaml:"auth_rid" env:"AUTH_RID"`
	SecondValue string `env-required:"true" yaml:"second-value" env:"SECOND"`
	User        string `yaml:"userr" env:"USERR" env-default:"user"`
	LocalSecret string `yaml:"local_secret" env:"LOCAL_SECRET" env-default:"local_secret"`
	ImapHost    string `yaml:"imap_host" env:"IMAP_HOST" env-default:"imap_host"`
}

type ConfigTestSuite struct {
	suite.Suite
	testStruct *testConfigStruct
	configEnv  map[string]string
	testYaml   string
}

func (s *ConfigTestSuite) SetupSuite() { /*before all test*/
	defaultConfigPath = "../tests/config_test.yaml"
	// defaultLocalConfigPath = "../tests/config_test.local.yaml"
	for i := range s.configEnv {
		os.Unsetenv(i)
	}

	f, err := os.OpenFile(defaultConfigPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	s.Require().NoError(err, "error when open file test.yaml")
	_, err = f.WriteString(strings.ReplaceAll(s.testYaml, "test", "test2"))
	s.Require().NoError(err, "error when write to  file test.yaml")

	defer f.Close()

	f1, err := os.OpenFile(defaultConfigPath, os.O_RDWR|os.O_CREATE, 0755)
	s.Require().NoError(err, "error when open file ", defaultConfigPath)
	_, err = f1.WriteString(s.testYaml)
	s.Require().NoErrorf(err, "error when write to file  %s", defaultConfigPath)
}

func (s *ConfigTestSuite) TearDownSuite() { /* after all test */
	defaultConfigPath = "./config/config_test.yaml"
	os.Remove(defaultConfigPath)
	os.Remove("./config/test.yaml")
}

func (s *ConfigTestSuite) SetupTest() { /*before each test*/
	for i, v := range s.configEnv {
		err := os.Setenv(i, v)
		s.Require().NoError(err)
	}
}

func (s *ConfigTestSuite) TearDownTest() { /* after each */
	for i := range s.configEnv {
		os.Unsetenv(i)
	}
}

func (s *ConfigTestSuite) TestFirst() {
	err := LoadConfig(s.testStruct)
	s.Require().NoError(err)
	s.Assert().Equal(s.testStruct.AuthKey1, "test from env")
	s.Assert().Equal(s.testStruct.SecondValue, "second-from file")
	s.Assert().Equal(s.testStruct.User, "user")
}

func (s *ConfigTestSuite) TestSecond() {
	os.Setenv("CONFIG", "./config/test.yaml")

	err := LoadConfig(s.testStruct)
	s.Require().NoError(err)
	s.Assert().Equal(s.testStruct.AuthKey1, "test from env")
	s.Assert().Equal(s.testStruct.SecondValue, "second-from file")
	s.Assert().Equal(s.testStruct.AuthRID, "test2FromFile")
	s.Assert().Equal(s.testStruct.User, "user")
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, &ConfigTestSuite{
		testStruct: &testConfigStruct{},
		configEnv:  map[string]string{"AUTH_KEY": "test from env"},
		testYaml: `---
auth_listen: ":3001" 
auth_rid: "testFromFile"
auth_key: "testFromFILE"
second-value: "second-from file"
`,
	})
}

func TestLoadRids(t *testing.T) {
	os.Setenv("CONFIG", "./config.yaml")
	c := &testConfigStruct{}
	err := LoadConfig(c)
	assert.NoError(t, err)
	fmt.Printf("%#v", c)
	assert.NotEmpty(t, c.AuthRID)
}

func TestLoadMulti(t *testing.T) {
	m := &struct {
		Env         string `env-required:"true"  yaml:"env" env:"ENV"`
		LocalSecret string `yaml:"local_secret" env:"LOCAL_SECRET"`
		ImapHost    string `yaml:"imap_host" env:"IMAP_HOST"`
	}{}
	os.Setenv("CONFIG", "../tests/config_test.yaml")
	err := LoadConfig(m)
	assert.NoError(t, err)
	fmt.Printf("%#v", m)
}
