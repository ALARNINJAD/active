package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	configFile = "sub.txt"
	maxConfigs = 50
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Paste v2ray configs (end with empty line):")
	var inputConfigs []string
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		inputConfigs = append(inputConfigs, line)
	}

	if len(inputConfigs) == 0 {
		fmt.Println("No configs provided.")
		return
	}

	fmt.Print("Enter visual name: ")
	baseName, _ := reader.ReadString('\n')
	baseName = strings.TrimSpace(baseName)
	if baseName == "" {
		fmt.Println("Invalid name.")
		return
	}

	existingConfigs := readConfigsFromFile(configFile)

	existingMap := make(map[string]string)
	for _, c := range existingConfigs {
		key := normalizeConfig(c)
		existingMap[key] = c
	}

	var newConfigs []string
	for i, c := range inputConfigs {
		key := normalizeConfig(c)

		delete(existingMap, key)

		newName := fmt.Sprintf("%s-%d", baseName, i+1)
		newConfigs = append(newConfigs, setVisualName(c, newName))
	}

	var finalConfigs []string

	finalConfigs = append(finalConfigs, newConfigs...)

	for _, c := range existingMap {
		finalConfigs = append(finalConfigs, c)
	}

	if len(finalConfigs) > maxConfigs {
		finalConfigs = finalConfigs[:maxConfigs]
	}

	writeConfigsToFile(configFile, finalConfigs)

	fmt.Println("Done. Configs updated successfully.")

	if err := AddCommitPush("update", "origin", "main"); err != nil {
		panic(err)
	}

}

func normalizeConfig(c string) string {
	if idx := strings.Index(c, "#"); idx != -1 {
		return c[:idx]
	}
	return c
}

func setVisualName(c, name string) string {
	base := normalizeConfig(c)
	return base + "#" + name
}

func readConfigsFromFile(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	var configs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			configs = append(configs, line)
		}
	}
	return configs
}

func writeConfigsToFile(path string, configs []string) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, c := range configs {
		writer.WriteString(c + "\n")
	}
	writer.Flush()
}

func AddCommitPush(commitMessage, remote, branch string) error {
	commands := [][]string{
		{"git", "add", "."},
		{"git", "commit", "-m", commitMessage},
		{"git", "push", remote, branch},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed (%v): %w", args, err)
		}
	}
	return nil
}
