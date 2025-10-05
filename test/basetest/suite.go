package basetest

import (
	"github.com/stretchr/testify/suite"
)

type Suite struct {
	suite.Suite
}

func (s *Suite) SetupSuite() {
}

func (s *Suite) BeforeTest(suiteName, testName string) {}

func (s *Suite) AfterTest(suiteName, testName string) {
}
