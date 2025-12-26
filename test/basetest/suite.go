package basetest

import (
	"github.com/stretchr/testify/suite"
	"github.com/tarmalonchik/golibs/config"
	kafka "github.com/tarmalonchik/golibs/kafkawrapper"
)

type Suite struct {
	suite.Suite

	Kafka kafka.Client
}

func (s *Suite) SetupSuite() {}

func (s *Suite) BeforeTest(suiteName, testName string) {
	cfg := kafka.Config{}

	err := config.Load(&cfg, "")
	s.Require().NoError(err)

	s.Kafka, err = kafka.NewClient(cfg, nil)
	s.Require().NoError(err)
}

func (s *Suite) AfterTest(suiteName, testName string) {}
