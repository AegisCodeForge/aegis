package gitlib

import (
	"os"
	"path"
	"strings"
)

// 1.  check GIT_COMMON_DIR_ENVIRONMENT
//     if there is such envvar, then the common dir is that value.
// 2.  else, check if file "{gitdir}/commondir" exists.
//     if no, then the common dir is "{gitdir}".
// 3.  else, read the content of "{gitdir}/commondir".
//     if the content is an absolute path, then it's the common dir.
//     else, then the common dir is "{gitdir}/{content}".
func getCommonDir(p string) string {
	envres := strings.TrimSpace(os.Getenv("GIT_COMMON_DIR_ENVIRONMENT"))
	if len(envres) > 0 { return envres }
	s, err := os.ReadFile(path.Join(p, "commondir"))
	if os.IsNotExist(err) { return p }
	sstr := string(s)
	if path.IsAbs(sstr) { return sstr }
	return path.Join(p, sstr)
}

// the checking process goes as follows. for any directory path `p`,
// any one of these must be valid for it to be considered as a valid
// git directory.
// 1.  check if the path "{p}/HEAD" is a valid HEAD reference, which
//     is always either one of these two:
//     + a symlink that redirects to a "refs/..." file;
//     + a normal text file, which contains either:
//       + a string that has the form of "ref: [name]" where "[name]"
//         should starts with "refs/";
//       + a valid sha-1 or sha-256 hash.
// 2.  check environment variable "DB_ENVIRONMENT". if it's set and
//     it's an accessible path, then it's considered a valid git
//     directory. the comment of git's code says it "checks
//     non-worktree-related signatures", whose meaning by the time of
//     writing this (2025.4.13) i haven't made sense of yet.
// for option 3 and 4 we need to get the "common dir" of "{p}", whose
// process is already listed above.
// 3.  a "refs" subdirecotry under the common dir, or:
// 4.  an "objects" subdirectory under the common dir.
func IsValidGitDirectory(p string) bool {
	if isValidHeadReference(path.Join(p, "HEAD")) { return true }
	if len(strings.TrimSpace(os.Getenv("DB_ENVIRONMENT"))) > 0 { return true }
	commondir := getCommonDir(p)
	_, err := os.ReadDir(path.Join(commondir, "refs"))
	if err == nil { return true }
	_, err = os.ReadDir(path.Join(commondir, "objects"))
	if err == nil { return true }
	return false
}

func isValidHeadReference(p string) bool {
	s, err := os.Readlink(p)
	if os.IsNotExist(err) { return false }
	if err == nil { return strings.HasPrefix(s, "refs/") }
	ss, err := os.ReadFile(p)
	if err != nil { return false }
	sstr := string(ss)
	if strings.HasPrefix(sstr, "ref:") {
		if strings.HasPrefix(
			strings.TrimSpace(sstr[len("ref:"):]),
			"refs/",
		) {
			return true
		}
	} else {
		if isValidHash(sstr) { return true }
	}
	return false
}

func isValidHash(s string) bool {
	ss := strings.TrimSpace(s)
	// either sha-1 or sha-256, which is 20 / 32 bytes or 40 / 64
	// characters in hex.
	if len(ss) != 40 && len(ss) != 64 { return false }
	for _, i := range ss {
		if !(('0' <= i && i <= '9') ||
			('a' <= i && i <= 'f') ||
			('A' <= i && i <= 'F')) { return false }
	}
	return true
}

