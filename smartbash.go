package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// struct for storing history
type Command struct {
	Text      string
	Frequency int
}

var commands []Command

func loadHistory() {
	// get path for bash history
	path := os.Getenv("HOME") + "/.bash_history"
	data, err := os.Open(path)
	if err != nil {
		return
	}
	defer data.Close()

	// read history file line by line
	freq := make(map[string]int)
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			freq[line]++
		}
	}

	// convert map to slice of Command structs
	for cmd, count := range freq {
		commands = append(commands, Command{Text: cmd, Frequency: count})
	}

	// sort commands by frequency
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Frequency > commands[j].Frequency
	})
}

func main() {
	loadHistory()

	fmt.Println("🧠 Smart Bash — your history suggestion")
	fmt.Println("Enter command or 'exit' to leave.")

}
