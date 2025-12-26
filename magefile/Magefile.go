package magefile

import (
	"flag"
	"fmt"
	"os"
	"strings"

	gosh "github.com/codeskyblue/go-sh"
	"github.com/magefile/mage/sh"
)

var (
	golangCiVersion        = "v2.5"
	isolationTestsFilePath = "test/cases"
)

func TestIsolation() {
	_, err := os.Stat(isolationTestsFilePath)
	if os.IsNotExist(err) {
		fmt.Println("No isolation test files found")
		return
	}

	if err := sh.Run("go", "clean", "--testcache"); err != nil {
		fmt.Println("Failed to run clean test cache", err)
		os.Exit(1)
	}

	out, err := sh.Output("go", "test", "-v", fmt.Sprintf("./%s/...", isolationTestsFilePath))
	if err != nil {
		fmt.Println("Failed to run isolation test", err)
		os.Exit(1)
	}
	fmt.Println(out)
}

func Test() {
	if err := sh.Run("go", "clean", "--testcache"); err != nil {
		fmt.Println("Failed to clean test cache")
		os.Exit(1)
	}

	pkg, err := sh.Output("bash", "-c", "go list ./... | grep -v test/")
	if err != nil {
		fmt.Println("Failed to get test package")
		os.Exit(1)
	}

	output, err := sh.Output("bash", "-c", "go test -v -race -coverprofile=coverage.out "+
		strings.ReplaceAll(pkg, "\n", " ")+" | tee test.log")
	if err != nil {
		fmt.Println("Failed to run tests")
		os.Exit(1)
	}
	fmt.Println(output)

	if output, err = sh.Output("bash", "-c", "go tool cover -func=coverage.out | grep total | awk '{print $3}'"); err != nil {
		fmt.Println("Failed to show total coverage")
		os.Exit(1)
	}
	fmt.Println(output)

	if err := os.Remove("coverage.out"); err != nil {
		fmt.Println("Failed to remove coverage.out")
	}
	if err := os.Remove("test.log"); err != nil {
		fmt.Println("Failed to remove test.log")
	}
	os.Exit(0)
}

func LintFix() error {
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		fmt.Println("Failed to run go mod tidy: %w", err)
		os.Exit(1)
	}

	if err := downloadLinter(); err != nil {
		fmt.Println("Downloading golangCi-lint failed: %w", err)
		os.Exit(1)
	}

	out, err := sh.Output(fmt.Sprintf("%s/bin/golangci-lint", PWD()), "run", "--timeout", "10m", "--verbose", "--fix")
	fmt.Println(out)
	return err
}

func Lint() error {
	if err := downloadLinter(); err != nil {
		fmt.Println("Downloading golangCi-lint failed: %w", err)
		os.Exit(1)
	}

	out, err := sh.Output(fmt.Sprintf("%s/bin/golangci-lint", PWD()), "run", "--timeout", "10m", "--verbose")
	fmt.Println(out)
	return err
}

func downloadLinter() error {
	err := sh.RunWith(
		map[string]string{
			"CGO_ENABLED": "0",
			"GOBIN":       fmt.Sprintf("%s/bin", PWD()),
		},
		"go",
		"install",
		"-v",
		fmt.Sprintf("github.com/golangci/golangci-lint/v2/cmd/golangci-lint@%s", golangCiVersion),
	)
	return err
}

// GitSync -force optional flag
func GitSync() {
	fs := flag.NewFlagSet("gitSync", flag.ContinueOnError)
	force := fs.Bool("force", false, "No confirmation")

	skipConfirm := false

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Printf("Failed to parse flags: %v\n", err)
		os.Exit(0)
	}
	if *force {
		skipConfirm = true
	}

	branch, err := sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		fmt.Printf("Failed to get git HEAD: %s\n", err)
		os.Exit(0)
	}
	fmt.Println(fmt.Sprintf("You are going to reset '$branch' to origin/%s", branch))

	if !skipConfirm {
		var cnf string
		fmt.Print("Confirm(y/n)?\n")

		if _, err := fmt.Scan(&cnf); err != nil {
			fmt.Printf("Failed to read confirmation: %v\n", err)
			os.Exit(0)
		}

		if cnf != "y" {
			fmt.Println("Aborting...")
			os.Exit(0)
		}

	} else {
		fmt.Println("Auto confirmed...")
	}

	if err := sh.Run("git", "remote", "prune", "origin"); err != nil {
		fmt.Println("Failed to prune origin", err)
		os.Exit(0)
	}

	if err := sh.Run("git", "fetch", "origin"); err != nil {
		fmt.Println("Failed to fetch origin", err)
		os.Exit(0)
	}

	if err := sh.Run("git", "reset", "--hard", fmt.Sprintf("origin/%s", branch)); err != nil {
		fmt.Println("Failed to reset --hard", err)
		os.Exit(0)
	}

	if err := sh.Run("git", "clean", "-fd"); err != nil {
		fmt.Println("Failed to clean local changes", err)
		os.Exit(0)
	}

	fmt.Println("Successful sync")
	os.Exit(0)
}

func KubectlConnect(kubeCtx, namespace, svc string, localPort, destPort uint32) *gosh.Session {
	sess := gosh.Command(
		"kubectl",
		fmt.Sprintf("--context=%s", kubeCtx),
		"port-forward",
		"-n",
		namespace,
		svc,
		fmt.Sprintf("%d:%d", localPort, destPort),
	)
	go func() {
		if err := sess.Run(); err != nil {
			fmt.Println("Could not connect to DB")
			os.Exit(0)
		}
	}()
	return sess
}

func PWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	return pwd
}
