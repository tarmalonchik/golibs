package basetest

import (
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type Suite struct {
	suite.Suite

	conn grpc.ClientConnInterface
}

func (s *Suite) SetupSuite() {
}

func (s *Suite) BeforeTest(suiteName, testName string) {}

func (s *Suite) AfterTest(suiteName, testName string) {
}
