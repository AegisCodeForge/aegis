package gitlib

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type MergeCheckConflictedFileInfo struct {
	// object mode
	Mode int `json:"mode"`
	ObjectId string `json:"oid"`
	Stage int `json:"stage"`
	FileName string `json:"fileName"`
}
type MergeCheckInformationalMessage struct {
	Path []string `json:"path"`
	Type string `json:"type"`
	Message string `json:"msg"`
}
type MergeCheckResult struct {
	Successful bool `json:"success"`
	ReceiverLocation string `json:"receiverLocation"`
	ReceiverBranch string `json:"receiverBranch"`
	ProviderRemoteName string `json:"providerRemoteName"`
	ProviderBranch string `json:"providerBranch"`
	ToplevelTreeOid string `json:"rootOid"`
	FileInfo []MergeCheckConflictedFileInfo `json:"fileInfo"`
	Message []MergeCheckInformationalMessage `json:"msg"`
}

func parseMergeCheckZResult(s string) *MergeCheckResult {
	// get top level oid.
	ss := strings.SplitN(s, "\x00", 2)
	topOid := string(ss[0])
	if len(ss[1]) <= 0 {
		return &MergeCheckResult{
			Successful: true,
			ToplevelTreeOid: topOid,
			FileInfo: nil,
			Message: nil,
		}
	}
	subj := ss[1]
	// get conflicted file info.
	cfl := make([]MergeCheckConflictedFileInfo, 0)
	for {
		ss := strings.SplitN(subj, "\x00", 2)
		if ss[0] == "" { subj = ss[1]; break }
		ss1 := strings.Split(ss[0], "\t")
		fileName := ss1[1]
		ss2 := strings.Split(ss1[0], " ")
		fileMode, _ := strconv.Atoi(ss2[0])
		fileObjId := ss2[1]
		stage, _ := strconv.Atoi(ss2[2])
		cfl = append(cfl, MergeCheckConflictedFileInfo{
			Mode: fileMode,
			ObjectId: fileObjId,
			Stage: stage,
			FileName: fileName,
		})
		subj = ss[1]
	}
	// get informational message.
	msgl := make([]MergeCheckInformationalMessage, 0)
	for len(subj) > 0 {
		// get list of paths
		pathList := make([]string, 0)
		ss = strings.SplitN(subj, "\x00", 2)
		subj = ss[1]
		pathnum, _ := strconv.Atoi(ss[0])
		for range pathnum {
			ss = strings.SplitN(subj, "\x00", 2)
			subj = ss[1]
			pathList = append(pathList, ss[0])
		}
		ss = strings.SplitN(subj, "\x00", 2)
		conflictType := ss[0]
		subj = ss[1]
		ss = strings.SplitN(subj, "\x00", 2)
		conflictMessage := ss[0]
		subj = ss[1]
		msgl = append(msgl, MergeCheckInformationalMessage{
			Path: pathList,
			Type: conflictType,
			Message: conflictMessage,
		})
	}
	return &MergeCheckResult{
		Successful: false,
		ToplevelTreeOid: topOid,
		FileInfo: cfl,
		Message: msgl,
	}
}

func (gr LocalGitRepository) SetUpMergeTarget(providerName string, providerPath string) error {
	cmd := exec.Command("git", "remote", "add", providerName, providerPath)
	cmd.Dir = gr.GitDirectoryPath
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		if strings.Contains(buf.String(), "already exists") { return nil }
		return err
	}
	return nil
}

func (gr LocalGitRepository) CheckBranchMergeConflict(localBranch string, remote string, remoteBranch string) (*MergeCheckResult, error) {
	// this would fetch the branch for you.
	cmd := exec.Command("git", "fetch", remote, remoteBranch)
	buf := new(bytes.Buffer)
	cmd.Stderr = buf
	cmd.Dir = gr.GitDirectoryPath
	err := cmd.Run()
	if err != nil {
		return nil, errors.New(err.Error() + ": " + buf.String())
	}
	remoteBranchFullName := fmt.Sprintf("%s/%s", remote, remoteBranch)
	cmd2 := exec.Command("git", "merge-tree", "-z", localBranch, remoteBranchFullName)
	buf.Reset()
	cmd2.Stdout = buf
	cmd2.Dir = gr.GitDirectoryPath
	err = cmd.Run()
	// NOTE: we must check exit status because that's what the
	// document says:
	//   Do NOT interpret an empty Conflicted file info
	//   list as a clean merge; check the exit status. A merge can have
	//   conflicts without having individual files conflict (there are a
	//   few types of directory rename conflicts that fall into this
	//   category, and others might also be added in the future).
	// the command would not put things to stderr - everything is put
	// to stdout.
	preres := parseMergeCheckZResult(buf.String())
	if err != nil {
		preres.Successful = false
	} else {
		preres.Successful = true
	}
	preres.ReceiverLocation = gr.GitDirectoryPath
	preres.ReceiverBranch = localBranch
	preres.ProviderRemoteName = remote
	preres.ProviderBranch = remoteBranch
	return preres, nil
}

