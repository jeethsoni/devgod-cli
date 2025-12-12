# Devgod

**Devgod** is a command-line tool that helps developers use Git more easily and confidently.

It guides you through common Git tasks so you don’t have to remember commands, naming rules, or best practices every time. You focus on writing code, and devgod helps handle the workflow around it.

It doesn’t replace Git or hide what’s happening. It simply makes each step clearer, safer, and easier to follow.

## 🦙 Install Ollama

Download and install Ollama from:

👉 <https://ollama.com>

or

```bash
brew install ollama
```

Pull a model:

```bash
ollama pull llama3.1
```

Verify Ollama is running:

```bash
ollama run llama3.1
```

## 📦 Installation

### macOS (recommended)

```sh
brew install jeethsoni/tap/devgod
```

## 🧠 What devgod does

### 🌿 Branch creation from intent

Instead of manually running:

```bash
git checkout -b what-should-i-name-this
```

you describe your intent:

```bash
dg git "add hello world script"
```

Devgod

- generates a branch name following best practices and naming conventions
- creates the branch
- checks you out to it automatically

This is ideal for developers who know what they want to build, but don’t want to think about branch naming.

## ✍️ Commit creation without guesswork

After making your changes, you run:

```bash
dg git
```

devgod:

- stages modified files automatically
- analyzes the staged changes
- proposes a commit message based on what actually changed
- shows a preview and asks for confirmation before committing

No more:

“I don’t know what to write for this commit message.”

## 🚀 Pull requests from the terminal

Once your work is committed, devgod can create a pull request directly from your terminal:

```bash
dg pr
```

devgod:

- asks which base branch to compare against
- lets you select reviewers interactively
- generates a pull request title and description using AI
- creates the pull request on GitHub after confirmation

This removes the need to switch to the browser just to open a PR.

## 🛣 Roadmap

- Cross-platform support (Windows & Linux)
- app scaffolding

## ⚠️ Known issues

Commit message suggestions may occasionally be inaccurate or not match your expectations.
Always review the proposed message before confirming the commit.

## 📄 License

MIT

Built with ❤️ by **Jeet Soni**

Happy Coding 💻
