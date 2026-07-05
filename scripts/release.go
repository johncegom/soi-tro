package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func runCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s %v failed: %w (stderr: %s)", name, args, err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func main() {
	// 1. Get the latest tag
	latestTag, err := runCmd("git", "describe", "--tags", "--abbrev=0")
	var nextTag string
	if err != nil {
		// No tag found, default to v0.1.0
		nextTag = "v0.1.0"
		fmt.Printf("No previous tag found. Starting with %s\n", nextTag)
	} else {
		// Parse latest tag (expecting format vX.Y.Z)
		re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
		matches := re.FindStringSubmatch(latestTag)
		if len(matches) != 4 {
			fmt.Printf("Latest tag %q does not match semantic versioning format (vX.Y.Z). Defaulting next tag to v0.1.0\n", latestTag)
			nextTag = "v0.1.0"
		} else {
			major, _ := strconv.Atoi(matches[1])
			minor, _ := strconv.Atoi(matches[2])
			patch, _ := strconv.Atoi(matches[3])

			// Check CLI arguments for minor or major version bump
			bump := "patch"
			if len(os.Args) > 1 {
				switch os.Args[1] {
				case "major":
					bump = "major"
					major++
					minor = 0
					patch = 0
				case "minor":
					bump = "minor"
					minor++
					patch = 0
				}
			}
			if bump == "patch" {
				patch++
			}
			nextTag = fmt.Sprintf("v%d.%d.%d", major, minor, patch)
			fmt.Printf("Latest tag: %s. Bumping %s version to: %s\n", latestTag, bump, nextTag)
		}
	}

	// 2. Create the tag locally
	fmt.Printf("Creating local git tag: %s\n", nextTag)
	_, err = runCmd("git", "tag", nextTag)
	if err != nil {
		fmt.Printf("Error creating tag: %v\n", err)
		os.Exit(1)
	}

	// 3. Push the tag to origin
	fmt.Printf("Pushing tag %s to origin...\n", nextTag)
	_, err = runCmd("git", "push", "origin", nextTag)
	if err != nil {
		fmt.Printf("Error pushing tag: %v\n", err)
		// Clean up the local tag if pushing failed
		_, _ = runCmd("git", "tag", "-d", nextTag)
		os.Exit(1)
	}

	fmt.Printf("Successfully pushed tag %s. GitHub Actions Release workflow will trigger shortly!\n", nextTag)
}
