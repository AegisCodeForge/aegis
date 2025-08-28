package model

import (
	"encoding/json"
	"strings"
)

type ACLTuple struct {
	AddMember bool `json:"addMember"`
	DeleteMember bool `json:"deleteMember"`
	EditMember bool `json:"editMember"`
	EditInfo bool `json:"editInfo"`
	AddRepository bool `json:"addRepo"`
	PushToRepository bool `json:"pushToRepo"`
	ArchiveRepository bool `json:"archiveRepo"`
	DeleteRepository bool `json:"deleteRepo"`
	EditHooks bool `json:"editHooks"`
}

type ACL struct {
	Version string `json:"version"`
	ACL map[string]*ACLTuple `json:"acl"`
}

func (aclt *ACLTuple) HasSettingPrivilege() bool {
	if aclt == nil { return false }
	// NOTE THAT having PushToRepository permission does not mean a
	// user is allowed to go into the setting panels of things.
	return aclt.AddMember || aclt.DeleteMember || aclt.EditMember || aclt.AddRepository || aclt.EditInfo || aclt.ArchiveRepository || aclt.DeleteRepository || aclt.EditHooks
}

func NewACL() *ACL {
	res := &ACL{
		Version: "0",
		ACL: make(map[string]*ACLTuple, 0),
	}
	return res
}

func ParseACL(s string) (*ACL, error) {
	if len(strings.TrimSpace(s)) <= 0 { return nil, nil }
	res := new(ACL)
	err := json.Unmarshal([]byte(s), res)
	if err != nil { return nil, err }
	return res, nil
}

func (at *ACLTuple) SerializeACLTuple() (string, error) {
	resbyte, err := json.MarshalIndent(at, "", "    ")
	if err != nil { return "", err }
	return string(resbyte), nil
}

func (s *ACL) SerializeACL() (string, error) {
	resbyte, err := json.MarshalIndent(s, "", "    ")
	if err != nil { return "", err }
	return string(resbyte), nil
}

func ToCommaSeparatedString(aclt *ACLTuple) string {
	if aclt == nil { return "" }
	res := make([]string, 0)
	if aclt.AddMember { res = append(res, "addMember") }
	if aclt.DeleteMember { res = append(res, "deleteMember") }
	if aclt.EditMember { res = append(res, "editMember") }
	if aclt.EditInfo { res = append(res, "editInfo") }
	if aclt.AddRepository { res = append(res, "addRepo") }
	if aclt.PushToRepository { res = append(res, "pushToRepo") }
	if aclt.ArchiveRepository { res = append(res, "archiveRepo") }
	if aclt.DeleteRepository { res = append(res, "deleteRepo") }
	if aclt.EditHooks { res = append(res, "editHooks") }
	return strings.Join(res, ",")
}

func (acl *ACL) GetUserPrivilege(username string) *ACLTuple {
	// return the tuple of whether the given user is a member.
	if acl == nil { return nil }
	r, e := acl.ACL[username]
	if !e { return nil } else { return r }
}


