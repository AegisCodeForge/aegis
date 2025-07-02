package model

import "github.com/bctnry/aegis/pkg/gitlib"

const (
	PULL_REQUEST_OPEN = 1
	PULL_REQUEST_CLOSED_AS_MERGED = 2
	PULL_REQUEST_CLOSED_AS_NOT_MERGED = 3
)

type PullRequest struct {
	PRId int64
	PRAbsId int64
	Timestamp int64
	ReceiverNamespace string
	ReceiverName string
	ReceiverBranch string
	ProviderNamespace string
	ProviderName string
	ProviderBranch string
	Status int
	MergeCheckResult *gitlib.MergeCheckResult
	MergeCheckTimestamp int64
}

const (
	PULL_REQUEST_EVENT_COMMENT = 1
	PULL_REQUEST_EVENT_COMMENT_ON_CODE = 2
	PULL_REQUEST_EVENT_UPDATE_ON_BRANCH = 3
	PULL_REQUEST_EVENT_MERGE_CONFLICT_CHECK = 4
	PULL_REQUEST_EVENT_CLOSE_AS_NOT_MERGED = 5
	PULL_REQUEST_EVENT_CLOSE_AS_MERGED = 6
)

type PullRequestEvent struct {
	PRAbsId int64
	// 1 - normal comment.
	// 2 - comment on code.
	// 3 - update on provider branch.
	// 4 - merge conflict check.
	// 5 - close as not merged.
	// 6 - close (merged).
	EventType int
	EventTimestamp int64
	EventAuthor string
	EventContent string
}

type PullRequestComment struct {
	Username string `json:"userName"`
	Content string `json:"content"`
}

type PullRequestCommentOnCode struct {
	CommitId string `json:"commitId"`
	Path string `json:"path"`
	LineRangeStart int `json:"lineRangeStart"`
	LineRangeEnd int `json:"lineRangeEnd"`
	Username string `json:"userName"`
	Content string `json:"content"`
}

type PullRequestUpdate struct {
	CommitId string `json:"commitId"`
}

type PullRequestMergeConflictCheck gitlib.MergeCheckResult

type PullRequestCloseAsNotMerged struct {
	Username string `json:"userName"`
}

type PullRequestCloseAsMerged struct {
	Username string `json:"userName"`
}

