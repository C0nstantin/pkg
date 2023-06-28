package sieve

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SieveClientTestSuite struct {
	suite.Suite
}

// ssh tunnel ssh 127.0.0.1:4190:4190  karataev@172.16.53.40
func (g *SieveClientTestSuite) TestSieveClientAll() {
	clt, err := NewSieveClient("127.0.0.1", "4190", "128756507061@wmkeeper.com", "---")
	defer clt.Close()
	g.Assert().NoError(err)
	_, name, err := clt.client.ListScripts()
	g.Assert().NoError(err)
	script, err := clt.client.GetScript(name)
	g.Assert().NoError(err)
	fmt.Printf("%s", script)
	warn, err := clt.client.CheckScript(script)
	g.Assert().NoError(err)
	g.Assert().Equal("", warn)
	warn1, err := clt.client.PutScript(name, script)
	g.Assert().NoError(err)
	g.Assert().Equal(warn1, "")

}

func TestSieveClientTest(t *testing.T) {
	suite.Run(t, &SieveClientTestSuite{})
}
