package model

import (
	"encoding/json"
	"strings"
)

type ACLTuple struct {
	AddMember bool
	DeleteMember bool
	EditMember bool
	AddRepository bool
	EditRepository bool
	PushToRepository bool
	ArchiveRepository bool
	DeleteRepository bool
}

type ACL struct {
	Version string
	ACL map[string]*ACLTuple
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
		addRepo := len(v) > 3 && v[3]
		editRepo := len(v) > 4 && v[4]
		pushToRepo := len(v) > 5 && v[5]
		archiveRepo := len(v) > 6 && v[6]
		deleteRepo := len(v) > 7 && v[7]
		res.ACL[k] = &ACLTuple{
			AddMember: addMember,
			DeleteMember: deleteMember,
			EditMember: editMember,
			AddRepository: addRepo,
			EditRepository: editRepo,
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
		vec := make([]bool, 6)
		vec[0] = v.AddMember
		vec[1] = v.DeleteMember
		vec[2] = v.EditMember
		vec[3] = v.AddRepository
		vec[4] = v.EditRepository
		vec[5] = v.PushToRepository
		vec[6] = v.ArchiveRepository
		vec[7] = v.DeleteRepository
		preres.ACL[k] = vec
	}
	resbyte, err := json.MarshalIndent(preres, "", "    ")
	if err != nil { return "", err }
	return string(resbyte), nil
}


