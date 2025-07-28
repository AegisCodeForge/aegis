package model

import (
	"errors"
	"fmt"
	"log"

	"github.com/bctnry/aegis/pkg/gitlib"
)

var ErrNotSupported = errors.New("Not supported type of local repository.")

func GetAegisType(x LocalRepository) uint8 {
	_, ok := x.(*gitlib.LocalGitRepository)
	if ok { return REPO_TYPE_GIT }
	return 0
}

func CreateLocalRepository(t uint8, namespace string, name string, p string) (LocalRepository, error) {
	switch t {
	case REPO_TYPE_GIT:
		return gitlib.NewLocalGitRepository(namespace, name, p), nil
	default:
		return nil, ErrNotSupported
	}
}

func CreateLocalForkOf(origin LocalRepository, newNs string, newName string, newP string) (LocalRepository, error) {
	switch GetAegisType(origin) {
	case REPO_TYPE_GIT:
		return origin.(*gitlib.LocalGitRepository).LocalForkTo(fmt.Sprintf("%s/%s", newNs, newName), newP), nil
	default:
		return nil, ErrNotSupported
	}
}

func GetLocalRepositoryLocalPath(r LocalRepository) string {
	switch GetAegisType(r) {
	case REPO_TYPE_GIT:
		return r.(*gitlib.LocalGitRepository).GitDirectoryPath
	default:
		log.Panic(ErrNotSupported)
	}
	return ""
}
