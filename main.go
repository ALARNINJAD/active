package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	subFile    = "sub.txt"
	maxConfigs = 100
	commitMsg  = "update v2ray subscription"
)

func normalizeConfig(cfg string) string {
	if i := strings.Index(cfg, "#"); i != -1 {
		return cfg[:i]
	}
	return cfg
}

func extractName(cfg string) string {
	if i := strings.Index(cfg, "#"); i != -1 {
		return cfg[i+1:]
	}
	return ""
}

func replaceName(cfg, newName string) string {
	base := normalizeConfig(cfg)
	return base + "#" + newName
}

func maxIndexForPrefix(configs []string, prefix string) int {
	re := regexp.MustCompile("^" + regexp.QuoteMeta(prefix) + "-(\\d+)$")
	max := 0
	for _, c := range configs {
		name := extractName(c)
		if m := re.FindStringSubmatch(name); len(m) == 2 {
			var n int
			fmt.Sscanf(m[1], "%d", &n)
			if n > max {
				max = n
			}
		}
	}
	return max
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("base name: ")
	baseName, _ := reader.ReadString('\n')
	baseName = strings.TrimSpace(baseName)

	fmt.Println("paste configs (end with CTRL+D):")

	var inputConfigs []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			inputConfigs = append(inputConfigs, line)
		}
	}

	var existing []string
	if data, err := os.ReadFile(subFile); err == nil {
		for _, l := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(l) != "" {
				existing = append(existing, strings.TrimSpace(l))
			}
		}
	}

	existingMap := make(map[string]string)
	for _, c := range existing {
		existingMap[normalizeConfig(c)] = c
	}

	startIndex := maxIndexForPrefix(existing, baseName) + 1

	var newConfigs []string
	for i, c := range inputConfigs {
		norm := normalizeConfig(c)

		delete(existingMap, norm)

		newName := fmt.Sprintf("%s-%d", baseName, startIndex+i)
		newConfigs = append(newConfigs, replaceName(c, newName))
	}

	var final []string
	final = append(final, newConfigs...)

	for _, c := range existingMap {
		final = append(final, c)
	}

	if len(final) > maxConfigs {
		final = final[:maxConfigs]
	}

	var buf bytes.Buffer
	for _, c := range final {
		buf.WriteString(c + "\n")
	}

	if err := os.WriteFile(subFile, buf.Bytes(), 0644); err != nil {
		fmt.Println("write error:", err)
		return
	}

	AddCommitPush()

	exec.Command("git", "add", "--all")
	exec.Command("git", "commit", "-m", "'update'")
	exec.Command("git", "push")
	fmt.Println("done ✅")
}

func AddCommitPush() error {
	commands := [][]string{
		{"git", "add", "--all"},
		{"git", "commit", "-m", "update"},
		{"git", "push"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed (%v): %w", args, err)
		}
	}
	return nil
}
