package model

import (
	"encoding/json"
	"strings"
)

type ACLTuple struct {
	AddMember bool
	DeleteMember bool
	EditMember bool
	EditInfo bool
	AddRepository bool
	PushToRepository bool
	ArchiveRepository bool
	DeleteRepository bool
}

type ACL struct {
	Version string
	ACL map[string]*ACLTuple
}

func (aclt *ACLTuple) HasSettingPrivilege() bool {
	if aclt == nil { return false }
	// NOTE THAT having PushToRepository permission does not mean a
	// user is allowed to go into the setting panels of things.
	return aclt.AddMember || aclt.DeleteMember || aclt.EditMember || aclt.AddRepository || aclt.EditInfo || aclt.ArchiveRepository || aclt.DeleteRepository
}

func ParseACL(s string) (*ACL, error) {
	if len(strings.TrimSpace(s)) <= 0 { return nil, nil }
	preres := struct{
		Version string
		ACL map[string][]bool
	}{}
	err := json.Unmarshal([]byte(s), &preres)
	if err != nil { return nil, err }
	res := new(ACL)
	res.ACL = make(map[string]*ACLTuple, 0)
	for k, v := range preres.ACL {
		addMember := len(v) > 0 && v[0]
		deleteMember := len(v) > 1 && v[1]
		editMember := len(v) > 2 && v[2]
		editInfo := len(v) > 3 && v[3]
		addRepo := len(v) > 4 && v[4]
		pushToRepo := len(v) > 5 && v[5]
		archiveRepo := len(v) > 6 && v[6]
		deleteRepo := len(v) > 7 && v[7]
		res.ACL[k] = &ACLTuple{
			AddMember: addMember,
			DeleteMember: deleteMember,
			EditMember: editMember,
			EditInfo: editInfo,
			AddRepository: addRepo,
			PushToRepository: pushToRepo,
			ArchiveRepository: archiveRepo,
			DeleteRepository: deleteRepo,
		}
	}
	return res, nil
}

func (s *ACL) SerializeACL() (string, error) {
	preres := struct{
		Version string
		ACL map[string][]bool
	}{}
	preres.Version = s.Version
	preres.ACL = make(map[string][]bool, 0)
	for k, v := range s.ACL {
		if v == nil { continue }
		vec := make([]bool, 8)
		vec[0] = v.AddMember
		vec[1] = v.DeleteMember
		vec[2] = v.EditMember
		vec[3] = v.EditInfo
		vec[4] = v.AddRepository
		vec[5] = v.PushToRepository
		vec[6] = v.ArchiveRepository
		vec[7] = v.DeleteRepository
		preres.ACL[k] = vec
	}
	resbyte, err := json.MarshalIndent(preres, "", "    ")
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
	return strings.Join(res, ",")
}

func (acl *ACL) GetUserPrivilege(username string) *ACLTuple {
	// return the tuple of whether the given user is a member.
	if acl == nil { return nil }
	r, e := acl.ACL[username]
	if !e { return nil } else { return r }
}


