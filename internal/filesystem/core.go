package filesystem

// Represents a contract of mechanism that returns
// single file path upon calling NextFilePath() method.
// If there are no more file paths available,
// this method returns an empty string.
type FilePathIterator interface {
	NextFilePath() string
}
