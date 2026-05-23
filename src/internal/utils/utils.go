package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func AskYesNo(question string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(question + " [y/N] ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	return input == "y" || input == "yes"
}

func AskPath(question string, defaultPath string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	if defaultPath != "" {
		question += " [" + defaultPath + "] "
	}

	fmt.Print(question + " ")

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" && defaultPath != "" {
		return defaultPath, nil
	} else if input == "" {
		return "", fmt.Errorf("no path provided")
	}

	return input, nil
}

func Ask(question string, defaultAnswer string, noEmptyAnswer bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	if defaultAnswer != "" {
		question += " [" + defaultAnswer + "]"
	}
	fmt.Print(question + " ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" && defaultAnswer != "" {
		return defaultAnswer, nil
	} else if input == "" && noEmptyAnswer {
		return "", fmt.Errorf("no answer provided")
	}
	return input, nil
}
