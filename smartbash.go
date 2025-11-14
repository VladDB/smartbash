package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/sahilm/fuzzy"
)

// struct for storing history
type Command struct {
	Text      string
	Frequency int
}

var (
	commands []Command
	p        *prompt.Prompt
)

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

func appendHistory(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}
	f, _ := os.OpenFile(os.Getenv("HOME")+"/.bash_history",
		os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if f != nil {
		defer f.Close()
		f.WriteString(cmd + "\n")
	}
}

func fuzzySearch(input string) []prompt.Suggest {
	var texts []string
	for _, c := range commands {
		texts = append(texts, c.Text)
	}

	matches := fuzzy.Find(input, texts)
	var suggestion []prompt.Suggest
	for _, m := range matches {
		suggestion = append(suggestion, prompt.Suggest{
			Text:        m.Str,
			Description: fmt.Sprintf("used %d times", commands[m.Index].Frequency),
		})
	}
	return suggestion
}

func completer(d prompt.Document) []prompt.Suggest {
	text := strings.TrimSpace(d.TextBeforeCursor())
	if text == "" {
		return nil
	}
	return fuzzySearch(text)
}

func executor(input string) {
	input = strings.TrimSpace(input)

	if input == "" {
		return
	}

	appendHistory(input)

	cmd := exec.Command("bash", "-c", input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

func exitChecker(in string, breakline bool) bool {
	trim := strings.TrimSpace(in)
	if trim == "exit" || trim == "quit" {
		return true
	}
	return false
}

func handleExit() {
	rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
	rawModeOff.Stdin = os.Stdin
	_ = rawModeOff.Run()
	rawModeOff.Wait()
}

func main() {
	defer handleExit()

	loadHistory()

	fmt.Println("🧠 Smart Bash — your history suggestion")
	fmt.Println("Enter command or 'exit' to leave.")

	p = prompt.New(
		executor,
		completer,
		prompt.OptionLivePrefix(func() (string, bool) {
			user := os.Getenv("USER")
			wd, _ := os.Getwd()
			host, _ := os.Hostname()
			home, _ := os.UserHomeDir()
			if home != "" && strings.HasPrefix(wd, home) {
				if rel, err := filepath.Rel(home, wd); err == nil {
					wd = filepath.Join("~", rel)
				}
			}
			return fmt.Sprintf("%s@%s:%s$", user, host, wd), true
		}),
		prompt.OptionSetExitCheckerOnInput(exitChecker),
		prompt.OptionTitle("Smart Bash Fuzzy"),
		prompt.OptionSuggestionBGColor(prompt.DarkBlue),
	)

	p.Run()
}
