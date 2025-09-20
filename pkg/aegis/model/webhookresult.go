package model

const (
	WEBHOOK_RESULT_UNDEFINED uint8 = 0
	WEBHOOK_RESULT_SUCCESS uint8 = 1
	WEBHOOK_RESULT_FAILURE uint8 = 2
)

type WebhookResult struct {
	Version uint8 `json:"ver"`
	UUID string `json:"id"`
	ReportUUID string `json:"rrUuid"`
	RepoNamespace string `json:"repoNs"`
	RepoName string `json:"repoName"`
	Status uint8 `json:"status"`
	Message string `json:"message"`
	Timestamp int64 `json:"timestamp"`
}

