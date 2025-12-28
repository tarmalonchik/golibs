//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/magefile/mage/sh"
	"github.com/tarmalonchik/golibs/magefile"
)

type Env string

const (
	Dev  = Env("dev")
	Prod = Env("prod")
)

func Bump() {
	loadEnv()

	magefile.Bump()
}

func LintFix() error {
	loadEnv()

	return magefile.LintFix()
}

func Lint() error {
	loadEnv()

	return magefile.Lint()
}

func InfraUp() {
	loadEnv()

	kafkaUp()
}

func InfraDown() {
	loadEnv()

	kafkaDown()
}

func TestIsolation() {
	loadEnv()

	magefile.TestIsolation()
}

func Test() {
	loadEnv()

	magefile.Test()
}

func kafkaUp() {
	origin := magefile.PWD()

	if err := os.Chdir("./deployment/kafka"); err != nil {
		fmt.Println("Could not change directory", err)
		os.Exit(1)
	}

	err := sh.RunWith(map[string]string{
		"KAFKA_PORT": os.Getenv("KAFKA_PORT"),
	},
		"docker",
		"compose",
		"up",
		"-d",
	)
	if err != nil {
		fmt.Println("While running kafka", err)
		os.Exit(1)
	}

	defer func() {
		if err := os.Chdir(origin); err != nil {
			fmt.Println("Could not roll back the dir", err)
		}
	}()
}

func kafkaDown() {
	origin := magefile.PWD()

	if err := os.Chdir("./deployment/kafka"); err != nil {
		fmt.Println("Could not change directory", err)
		os.Exit(1)
	}

	if err := sh.Run("docker", "compose", "down"); err != nil {
		fmt.Println("While running kafka", err)
		os.Exit(1)
	}

	defer func() {
		if err := os.Chdir(origin); err != nil {
			fmt.Println("Could not roll back the dir", err)
		}
	}()
}

func loadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Could not load .env file", err)
		os.Exit(1)
	}
}
