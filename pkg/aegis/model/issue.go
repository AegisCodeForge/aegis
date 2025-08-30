package model

type Issue struct {
	IssueAbsId int64
	RepoNamespace string
	RepoName string
	IssueId int
	IssueAuthor string
	IssueTitle string
	IssueContent string
	IssueTime int64
	IssueStatus int
	IssuePriority int
}

const (
	ISSUE_OPENED = 1
	ISSUE_CLOSED_AS_SOLVED = 2
	ISSUE_CLOSED_AS_DISCARDED = 3
)

const (
	EVENT_COMMENT = 1
	EVENT_CLOSED_AS_SOLVED = 2
	EVENT_CLOSED_AS_DISCARDED = 3
	EVENT_REOPENED = 4
)

type IssueEvent struct {
	EventAbsId int64
	EventType int
	EventTimestamp int64
	EventAuthor string
	EventContent string	
}

