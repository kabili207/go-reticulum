package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var pure_python bool

const pure_notice = "\n\n**Warning!** *This package is the zero-dependency version of Reticulum. You should almost certainly use the [normal package](https://pypi.org/project/rns) instead. Do NOT install this package unless you know exactly why you are doing it!*"

type setupConfig struct {
	name                  string
	version               string
	author                string
	author_email          string
	description           string
	long_description      string
	long_description_type string
	url                   string
	packages              []string
	license               string
	license_files         []string
	classifiers           []string
	entry_points          map[string][]string
	install_requires      []string
	python_requires       string
	excluded_modules      []string
	raw_arguments         []string
}

func readVersion(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	re := regexp.MustCompile(`__version__\s*=\s*["']([^"']+)["']`)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if m := re.FindStringSubmatch(line); m != nil {
			return m[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("version not found in %s", path)
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Rough equivalent of setuptools.setup; currently just prints derived config.
func setup(cfg setupConfig) {
	fmt.Printf("Building package %s (%s)\n", cfg.name, cfg.version)
	fmt.Printf("Pure python: %v\n", pure_python)
	fmt.Printf("Requirements: %v\n", cfg.install_requires)
	fmt.Printf("Entry points: %v\n", cfg.entry_points)
	// Add real build/file-writing logic here if needed.
}

func main() {
	// Keep the same API: the --pure flag.
	flag.BoolVar(&pure_python, "pure", false, "build pure-python wheel")
	flag.Parse()

	if pure_python {
		fmt.Println("Building pure-python wheel")
	}

	version, err := readVersion(filepath.Join("RNS", "_version.py"))
	if err != nil {
		log.Fatal(err)
	}

	long_description, err := readFile("README.md")
	if err != nil {
		log.Fatal(err)
	}

	var pkg_name string
	var requirements []string

	if pure_python {
		pkg_name = "rnspure"
		requirements = []string{}
		long_description = strings.ReplaceAll(long_description, "</p>", "</p>"+pure_notice)
	} else {
		pkg_name = "rns"
		requirements = []string{"cryptography>=3.4.7", "pyserial>=3.5"}
	}

	excluded_modules := []string{"tests.*", "tests"}

	// There is no direct find_packages equivalent; provide a package list manually.
	packages := []string{
		"RNS",
		// Add other packages here if needed.
	}

	entry_points := map[string][]string{
		"console_scripts": {
			"rnsd=RNS.Utilities.rnsd:main",
			"rnstatus=RNS.Utilities.rnstatus:main",
			"rnprobe=RNS.Utilities.rnprobe:main",
			"rnpath=RNS.Utilities.rnpath:main",
			"rnid=RNS.Utilities.rnid:main",
			"rncp=RNS.Utilities.rncp:main",
			"rnx=RNS.Utilities.rnx:main",
			"rnir=RNS.Utilities.rnir:main",
			"rnodeconf=RNS.Utilities.rnodeconf:main",
		},
	}

	cfg := setupConfig{
		name:                  pkg_name,
		version:               version,
		author:                "Mark Qvist",
		author_email:          "mark@unsigned.io",
		description:           "Self-configuring, encrypted and resilient mesh networking stack for LoRa, packet radio, WiFi and everything in between",
		long_description:      long_description,
		long_description_type: "text/markdown",
		url:                   "https://reticulum.network/",
		packages:              packages,
		license:               "Reticulum License",
		license_files:         []string{"LICENSE"},
		classifiers: []string{
			"Programming Language :: Python :: 3",
			"Operating System :: OS Independent",
			"Development Status :: 5 - Production/Stable",
		},
		entry_points:     entry_points,
		install_requires: requirements,
		python_requires:  ">=3.7",
		excluded_modules: excluded_modules,
		raw_arguments:    os.Args[1:],
	}

	setup(cfg)
}
