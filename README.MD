# Smart Bash

**Smart Bash** is an interactive command-line shell that provides advanced autocompletion capabilities and intelligent Bash history search.
The application is designed to improve terminal productivity by leveraging fuzzy search, dynamic line analysis, and filesystem path completion.

---

## Key Features

### 1. Filesystem Path Autocompletion

Smart Bash automatically suggests path candidates when a token contains:

* `/`
* `~`
* `.`

Features include:

* home directory expansion (`~/`);
* support for both relative and absolute paths;
* automatic `/` suffix for directories;
* directory content inspection;
* completion generated relative to the full command line prefix.

### 2. Fuzzy Search Through Bash History

The application loads command history from `~/.bash_history` and builds a frequency-based index.

Capabilities:

* fuzzy matching across the entire command history;
* suggestions sorted by match quality and usage frequency;
* automatic addition of executed commands to history;
* contextual insertion based on the current cursor position.

### 3. Extended `cd` Command

The `cd` command is implemented as a built-in:

* default transition to the home directory when no argument is provided;
* support for paths starting with `~`;
* error handling for invalid directories.

### 4. Dynamic Prompt Prefix

The prompt displays user, hostname, and working directory in the following format:

```
user@host:~/relative/path$
```

Paths inside the home directory are automatically shortened to `~`.

### 5. Terminal State Restoration

Upon exit, Smart Bash restores the terminal from raw mode to its normal state.

---

## Installation

### Build from source

Go ≥ 1.20 is required.

```bash
go build -o smartbash smartbash.go
```

### Build using Makefile

A Makefile is included in the project and provides build targets for multiple platforms, including Linux and macOS.
All release binaries are built using this Makefile.

Command to get **.deb** file for installing into your system,
```bash
make deb
```

or for getting execute file.
```bash
make build
```

### Download prebuilt releases

Precompiled binaries are available in the repository’s [Releases](https://github.com/VladDB/smartbash/releases) section.
These releases are packaged automatically according to the Makefile rules.

---

## Usage

Run the application:

```bash
./smartbash
```

To exit, use:

```
exit
quit
```

---

## Dependencies

Smart Bash makes use of the following external libraries:

* **go-prompt** — interactive CLI engine and suggestion system
* **fuzzy** — fuzzy-matching implementation

---
