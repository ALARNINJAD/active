package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const maxLines = 100
const fileName = "sub.txt"

func main() {
	reader := bufio.NewReader(os.Stdin)

	// 1. read configs from terminal
	fmt.Println("Paste configs (empty line to finish):")
	var inputConfigs []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		inputConfigs = append(inputConfigs, line)
	}

	// 2. read name
	fmt.Print("Enter name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	// 3. read file
	fileConfigs := readLines(fileName)

	// 4. build map of input configs (without remark)
	inputMap := make(map[string]bool)
	for _, c := range inputConfigs {
		inputMap[stripRemark(c)] = true
	}

	// 5. remove duplicates from file
	var filtered []string
	for _, c := range fileConfigs {
		if !inputMap[stripRemark(c)] {
			filtered = append(filtered, c)
		}
	}

	// 6. rename remaining file configs to old-*
	var oldConfigs []string
	for i, c := range filtered {
		oldConfigs = append(oldConfigs, addRemark(c, fmt.Sprintf("old-%d", i+1)))
	}

	// 7. rename input configs and prepend
	var newConfigs []string
	for i, c := range inputConfigs {
		newConfigs = append(newConfigs, addRemark(c, fmt.Sprintf("%s-%d", name, i+1)))
	}

	finalConfigs := append(newConfigs, oldConfigs...)

	// 8. trim to 100
	if len(finalConfigs) > maxLines {
		finalConfigs = finalConfigs[:maxLines]
	}

	// write file
	writeLines(fileName, finalConfigs)

	// 9. git commands
	run("git", "add", fileName)
	run("git", "commit", "-m", "update")
	run("git", "push")
}

// ---------- helpers ----------

func stripRemark(c string) string {
	if i := strings.Index(c, "#"); i != -1 {
		return c[:i]
	}
	return c
}

func addRemark(c, remark string) string {
	base := stripRemark(c)
	return base + "#" + remark
}

func readLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func writeLines(path string, lines []string) {
	file, _ := os.Create(path)
	defer file.Close()

	for _, l := range lines {
		fmt.Fprintln(file, l)
	}
}

func run(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	_ = c.Run()
}
