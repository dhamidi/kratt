package worker

// GitHub interface encapsulates GitHub operations
type GitHub interface {
	// GetPRInfo retrieves pull request information including comments
	GetPRInfo(prNumber int) (string, error)

	// PostComment posts a comment to the specified pull request
	PostComment(prNumber int, body string) error
}
