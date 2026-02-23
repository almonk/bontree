package ui

// gitFileStatus represents the git state of a file
type gitFileStatus int

const (
	gitUnchanged  gitFileStatus = iota
	gitModified                         // working tree modified
	gitAdded                            // staged / new tracked file
	gitDeleted                          // deleted
	gitUntracked                        // untracked (?)
	gitIgnored                          // ignored (!)
)

type clearFlashMsg struct{}
