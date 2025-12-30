package magefile

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/sh"
)

var (
	golangCiVersion        = "v2.5"
	isolationTestsFilePath = "test/cases"
)

type Env string

const (
	Dev  = Env("dev")
	Prod = Env("prod")
)

func Bump() {
	cmd := `bash <(curl -fsSL https://raw.githubusercontent.com/tarmalonchik/golibs/main/scripts/git_increment_tag.bash)`

	if err := sh.Run("bash", "-c", cmd); err != nil {
		fmt.Println("Failed to run bump", err)
		os.Exit(1) //nolint:revive
	}
	os.Exit(0) //nolint:revive
}

func TestIsolation() {
	_, err := os.Stat(isolationTestsFilePath)
	if os.IsNotExist(err) {
		fmt.Println("No isolation test files found")
		return
	}

	if err := sh.Run("go", "clean", "--testcache"); err != nil {
		fmt.Println("Failed to run clean test cache", err)
		os.Exit(1) //nolint:revive
	}

	out, err := sh.Output("go", "test", "-v", fmt.Sprintf("./%s/...", isolationTestsFilePath))
	if err != nil {
		fmt.Println(out)
		fmt.Println("Failed to run isolation test", err)
		os.Exit(1) //nolint:revive
	}
	fmt.Println(out)
}

func Test() {
	if err := sh.Run("go", "clean", "--testcache"); err != nil {
		fmt.Println("Failed to clean test cache")
		os.Exit(1) //nolint:revive
	}

	pkg, err := sh.Output("bash", "-c", "go list ./... | grep -v test/")
	if err != nil {
		fmt.Println("Failed to get test package")
		os.Exit(1) //nolint:revive
	}

	output, err := sh.Output("bash", "-c", "go test -v -race -coverprofile=coverage.out "+
		strings.ReplaceAll(pkg, "\n", " ")+" | tee test.log")
	if err != nil {
		fmt.Println(output)
		fmt.Println("Failed to run tests")
		os.Exit(1) //nolint:revive
	}
	fmt.Println(output)

	if output, err = sh.Output("bash", "-c", "go tool cover -func=coverage.out | grep total | awk '{print $3}'"); err != nil {
		fmt.Println("Failed to show total coverage")
		fmt.Println(output)
		os.Exit(1) //nolint:revive
	}
	fmt.Println(output)

	if err := os.Remove("coverage.out"); err != nil {
		fmt.Println("Failed to remove coverage.out")
	}
	if err := os.Remove("test.log"); err != nil {
		fmt.Println("Failed to remove test.log")
	}
	os.Exit(0) //nolint:revive
}

func LintFix() error {
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		fmt.Println("Failed to run go mod tidy: %w", err)
		os.Exit(1) //nolint:revive
	}

	if err := downloadLinter(); err != nil {
		fmt.Println("Downloading golangCi-lint failed: %w", err)
		os.Exit(1) //nolint:revive
	}

	out, err := sh.Output(fmt.Sprintf("%s/bin/golangci-lint", PWD()), "run", "--timeout", "10m", "--verbose", "--fix")
	fmt.Println(out)
	return err
}

func Generate() {
	const protoPath = "./proto"

	origin := PWD()

	_, err := os.Stat(protoPath)
	if os.IsNotExist(err) {
		fmt.Println("No proto path found")
		return
	}

	if err := os.Chdir(protoPath); err != nil {
		fmt.Println("Could not change directory", err)
		return
	}

	cmd := `bash <(curl -fsSL https://raw.githubusercontent.com/tarmalonchik/golibs/main/scripts/gen_proto.bash)`

	if err := sh.Run("bash", "-c", cmd); err != nil {
		fmt.Println("Failed to run proto generate", err)
		return
	}

	defer func() {
		if err := os.Chdir(origin); err != nil {
			fmt.Println("Could not roll back the dir", err)
		}
	}()
}

func Lint() error {
	if err := downloadLinter(); err != nil {
		fmt.Println("Downloading golangCi-lint failed: %w", err)
		os.Exit(1) //nolint:revive
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
		os.Exit(0) //nolint:revive
	}
	if *force {
		skipConfirm = true
	}

	branch, err := sh.Output("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		fmt.Printf("Failed to get git HEAD: %s\n", err)
		os.Exit(0) //nolint:revive
	}
	fmt.Printf("You are going to reset '$branch' to origin/%s\n", branch)

	if !skipConfirm {
		var cnf string
		fmt.Print("Confirm(y/n)?\n")

		if _, err := fmt.Scan(&cnf); err != nil {
			fmt.Printf("Failed to read confirmation: %v\n", err)
			os.Exit(0) //nolint:revive
		}

		if cnf != "y" {
			fmt.Println("Aborting...")
			os.Exit(0) //nolint:revive
		}
	} else {
		fmt.Println("Auto confirmed...")
	}

	if err := sh.Run("git", "remote", "prune", "origin"); err != nil {
		fmt.Println("Failed to prune origin", err)
		os.Exit(0) //nolint:revive
	}

	if err := sh.Run("git", "fetch", "origin"); err != nil {
		fmt.Println("Failed to fetch origin", err)
		os.Exit(0) //nolint:revive
	}

	if err := sh.Run("git", "reset", "--hard", fmt.Sprintf("origin/%s", branch)); err != nil {
		fmt.Println("Failed to reset --hard", err)
		os.Exit(0) //nolint:revive
	}

	if err := sh.Run("git", "clean", "-fd"); err != nil {
		fmt.Println("Failed to clean local changes", err)
		os.Exit(0) //nolint:revive
	}

	fmt.Println("Successful sync")
	os.Exit(0) //nolint:revive
}

func KubectlConnect(ctx context.Context, kubeCtx, namespace, svc string, localPort, destPort uint32) error {
	kubectlCmd := fmt.Sprintf("kubectl --context=%s port-forward -n %s %s %d:%d",
		kubeCtx, namespace, svc, localPort, destPort)

	cmd := exec.CommandContext(ctx, "zsh", "-i", "-c", kubectlCmd)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func PWD() string {
	pwd, err := os.Getwd()
	if err != nil {
		os.Exit(1) //nolint:revive
	}
	return pwd
}
