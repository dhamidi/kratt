package worker

// LocalGit interface encapsulates git worktree operations
type LocalGit interface {
	// CheckWorktreeExists checks if a worktree exists for the given branch
	CheckWorktreeExists(branch string) (bool, error)

	// CreateWorktree creates a new worktree for the given branch at the specified path
	CreateWorktree(branch, path string) error

	// ChangeDirectory changes to the specified worktree directory
	ChangeDirectory(path string) error

	// CommitAndPush commits all changes and pushes to the remote branch
	CommitAndPush(message string) error

	// GetWorktreePath returns the path to the worktree for the given branch
	GetWorktreePath(branch string) (string, error)
}
