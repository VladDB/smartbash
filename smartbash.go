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
	freqMap  = map[string]int{}
)

// ------------------------------------------
// HISTORY + CACHE
// ------------------------------------------

func loadHistory() {
	// get path for bash history
	path := filepath.Join(os.Getenv("HOME"), ".bash_history")
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		cmd := strings.TrimSpace(scanner.Text())
		if cmd == "" {
			continue
		}
		freqMap[cmd]++
	}

	rebuildCache()
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

	// append to cache
	freqMap[cmd]++
	rebuildCache()
}

func rebuildCache() {
	commands = commands[:0]
	for c, f := range freqMap {
		commands = append(commands, Command{Text: c, Frequency: f})
	}
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Frequency > commands[j].Frequency
	})
}

// ------------------------------------------
// PATH COMPLETION
// ------------------------------------------

// expands leading "~" to actual home dir
func expandHome(p string) string {
	if strings.HasPrefix(p, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

// build suggestions for paths. `rawToken` is what user currently typed for the path (may contain ~).
// `prefixBeforeToken` is the part of the command before the token (e.g. "cd " or "git add ").
func pathSuggestions(prefixBeforeToken, rawToken string) []prompt.Suggest {
	expanded := expandHome(rawToken)
	// if empty token -> list current dir
	if expanded == "" {
		expanded = "."
	}

	dir := expanded

	// if path ends with slash, we list that directory contents
	if strings.HasSuffix(expanded, string(os.PathSeparator)) {
		dir = expanded
	} else {
		dir = filepath.Dir(expanded)
	}

	// if dir is empty or ".", set to "."
	if dir == "" {
		dir = "."
	}

	// try to read directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		// return nothing if can't read
		return nil
	}

	// extract the filter prefix (characters after last /)
	var filterPrefix string
	if !strings.HasSuffix(expanded, string(os.PathSeparator)) {
		filterPrefix = filepath.Base(expanded)
	}

	var suggests []prompt.Suggest
	for _, e := range entries {
		name := e.Name()

		// filter by prefix if it exists
		if filterPrefix != "" && !strings.HasPrefix(name, filterPrefix) {
			continue
		}

		full := filepath.Join(dir, name)

		// restore leading ~ if original used it
		display := full
		if strings.HasPrefix(rawToken, "~") {
			if home, err := os.UserHomeDir(); err == nil {
				if strings.HasPrefix(full, home) {
					display = filepath.Join("~", strings.TrimPrefix(full, home))
				}
			}
		}
		// if entry is dir -> append slash to hint
		if e.IsDir() {
			display = display + string(os.PathSeparator)
			full = full + string(os.PathSeparator)
		}
		// Build the full command line suggestion: prefixBeforeToken + display
		suggestText := strings.TrimRight(prefixBeforeToken, "") + display
		suggests = append(suggests, prompt.Suggest{Text: suggestText, Description: full})
	}

	return suggests
}

// returns the part of the line up to (but not including) the token that should be completed
// and the token itself.
func splitLineLastToken(line string) (prefixBeforeToken, token string) {
	// we will split by spaces, but preserve everything before the last token (including trailing spaces)
	if strings.TrimSpace(line) == "" {
		return "", ""
	}
	// find last space
	i := strings.LastIndex(line, " ")
	if i == -1 {
		return "", line
	}
	prefix := line[:i+1] // include the space
	token = line[i+1:]

	// If token is a path, extract path part from prefix
	if isPathToken(token) {
		// Find last space in prefix to get the actual path prefix
		parts := strings.Fields(prefix)
		if len(parts) > 0 && isPathToken(parts[len(parts)-1]) {
			// Last part of prefix is a path, keep it with space
			lastPathIdx := strings.LastIndex(prefix[:len(prefix)-1], " ")
			if lastPathIdx == -1 {
				return "", token
			}
			return prefix[lastPathIdx+1:], token
		}
		return "", token
	}
	return "", token
}

func isPathToken(tok string) bool {
	return strings.Contains(tok, "/") ||
		strings.HasPrefix(tok, "~") ||
		strings.HasPrefix(tok, ".")
}

// ------------------------------------------
// COMPLETER
// ------------------------------------------

func completer(d prompt.Document) []prompt.Suggest {

	line := d.TextBeforeCursor()
	if strings.TrimSpace(line) == "" {
		return nil
	}

	prefix, token := splitLineLastToken(line)
	trim := strings.TrimSpace(line)

	// GENERAL PATH COMPLETION (FOR ANY COMMAND)
	if isPathToken(token) {
		return pathSuggestions(prefix, token)
	}

	// FUZZY HISTORY COMPLETION
	return fuzzySearch(trim)
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

// ------------------------------------------
// EXECUTOR
// ------------------------------------------

func executor(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	if strings.HasPrefix(input, "cd") {
		parts := strings.Fields(input)
		var target string

		if len(parts) == 1 {
			home, _ := os.UserHomeDir()
			target = home
		} else {
			target = expandHome(parts[1])
		}

		if err := os.Chdir(target); err != nil {
			fmt.Println("cd:", err)
			return
		}

		appendHistory(input)
		return
	}

	appendHistory(input)

	cmd := exec.Command("bash", "-c", input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

func exitChecker(in string, breakLine bool) bool {
	trim := strings.TrimSpace(in)
	if trim == "exit" || trim == "quit" {
		return true
	}
	return false
}

func livePrefix() (string, bool) {
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

	p := prompt.New(
		executor,
		completer,
		prompt.OptionLivePrefix(livePrefix),
		prompt.OptionSetExitCheckerOnInput(exitChecker),
		prompt.OptionTitle("Smart Bash Fuzzy"),
		prompt.OptionSuggestionBGColor(prompt.DarkBlue),
	)

	p.Run()
}
