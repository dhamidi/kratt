# 🧙‍♂️ Kratt

> *Your magical helper for automated pull request processing*

Meet Kratt, your friendly AI-powered coding assistant that turns tedious PR work into magic! Just like the legendary treasure-bearing creatures from Estonian folklore, Kratt works tirelessly behind the scenes to help you with your code.

## What's a Kratt, anyway?

In Estonian folklore, a **kratt** is a magical creature crafted by its master from everyday household items like brooms, rakes, or wooden sticks. Once brought to life, this helpful spirit becomes a devoted servant, working day and night to bring treasure and assistance to its creator.

Our Kratt follows the same spirit—it's your personal coding assistant, built from modern tools (Go, GitHub CLI, and AI), that springs into action to help process pull requests, create branches, and handle the repetitive tasks that slow down your development workflow.

## Quick Start

### Installation

Make sure you have these prerequisites:
- Go 1.24 or later
- Git
- GitHub CLI (`gh`) installed and authenticated

Then install Kratt:

```bash
go install github.com/dhamidi/kratt@latest
```

### Your First Magic Spell

Process an existing pull request:
```bash
cd your-project
kratt worker run 42  # Process PR #42
```

Create a new feature branch with instructions:
```bash
kratt worker start feature/auth "Implement JWT authentication"
```

## How It Works

### `kratt worker run <pr-number>`

Your Kratt will:
1. 🕵️ Detect your git repository and GitHub remote
2. 📋 Analyze the specified pull request
3. 🤖 Run an AI agent to review and improve the code
4. ✅ Execute linting and tests
5. 📤 Push any improvements back to the PR

```bash
kratt worker run 1        # Process PR #1
kratt worker run 42       # Process PR #42
```

### `kratt worker start <branch-name> <instructions>`

Your Kratt will:
1. 🌿 Create a new branch with your specified name
2. 📝 Write your instructions to a markdown file
3. 💾 Commit and push the new branch
4. 🎯 Open a pull request ready for implementation

```bash
kratt worker start feature/auth "Implement JWT authentication"
kratt worker start bugfix/login-error "Fix login validation bug"
```

## Configuration

Want to customize your Kratt's behavior? Use these flags:

```bash
# Extend the timeout for complex tasks
kratt worker run 1 --timeout 45m

# Use a different AI agent
kratt worker run 1 --agent "claude-dev --stdin"

# Custom linting command
kratt worker run 1 --lint "golangci-lint run"

# Custom test command  
kratt worker run 1 --test "go test -v ./..."

# Use custom instructions
kratt worker run 1 --instructions ./my-instructions.txt
```

### Default Settings

Your Kratt comes pre-configured with sensible defaults:
- **Timeout**: 30 minutes
- **AI Agent**: `amp --stdin`
- **Linting**: `go fmt ./...`
- **Testing**: `go test ./...`

## Troubleshooting

**"Error: current directory is not a git repository"**
- Make sure you're in a git repository: `git status`

**"Error: no GitHub remote found in current repository"**
- Add a GitHub remote: `git remote add origin https://github.com/username/repo.git`

**"Error: invalid pull request number"**
- Use a positive integer: `kratt worker run 42` (not `kratt worker run abc`)

**GitHub API errors**
- Make sure GitHub CLI is authenticated: `gh auth status`
- Check your repository permissions

## Contributing

Found a bug? Want to teach your Kratt new tricks? Contributions are welcome!

1. Fork the repository
2. Create your feature branch: `kratt worker start feature/amazing-feature "Add amazing feature"`
3. Make your changes
4. Test your changes: `go test ./...`
5. Submit a pull request

## Development

### Building from source

```bash
git clone https://github.com/dhamidi/kratt.git
cd kratt
go build
```

### Running tests

```bash
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

---

*Just like the kratts of old Estonian folklore, may this tool bring you treasure in the form of cleaner code and smoother workflows! 🪄*
