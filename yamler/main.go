package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/geofffranks/simpleyaml"
	"github.com/geofffranks/spruce"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

//go:embed base.yml
var baseFile []byte

func main() {
	outputPath := flag.String("o", ".golangci.yml", "output file path")
	templatePath := flag.String("t", ".golangci-tmpl.yml", "template file path")

	flag.Usage = func() {
		fmt.Printf("usage: yamler [options]\n\n")
		fmt.Println()
		fmt.Printf("available params:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if outputPath == nil || templatePath == nil {
		log.Fatal("output file path or template file path is required")
	}

	if err := merge(*outputPath, *templatePath); err != nil {
		log.Fatal(err)
	}
}

func merge(outputPath, templatePath string) error {
	templateFile, err := os.ReadFile(templatePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read template file %s: %w", templatePath, err)
	}

	templateMap, err := getMapFrom(templateFile, true)
	if err != nil {
		return fmt.Errorf("get template map: %w", err)
	}
	if len(templateMap) == 0 {
		logrus.Info("template map is empty, using empty one")
	}

	baseMap, err := getMapFrom(baseFile, true)
	if err != nil {
		return fmt.Errorf("get base map: %w", err)
	}

	result, err := spruce.Merge(baseMap, templateMap)
	if err != nil {
		return fmt.Errorf("merge files: %w", err)
	}

	out, err := yaml.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	if err := os.WriteFile(outputPath, out, 0644); err != nil {
		return fmt.Errorf("write file %s: %w", outputPath, err)
	}
	return nil
}

func getMapFrom(data []byte, allowEmpty bool) (map[interface{}]interface{}, error) {
	if len(data) == 0 && allowEmpty {
		return map[interface{}]interface{}{}, nil
	}

	yml, err := simpleyaml.NewYaml(data)
	if err != nil {
		return nil, fmt.Errorf("parse yaml file: %w", err)
	}

	mp, err := yml.Map()
	if err != nil {
		return nil, fmt.Errorf("parse yaml to map: %w", err)
	}

	return mp, nil
}
