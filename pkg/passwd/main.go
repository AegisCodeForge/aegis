package passwd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PasswdUser struct {
	UserName string
	// 
	Password string
	UID int
	GID int
	Gecos string
	HomeDir string
	Shell string
}

type Passwd map[string]*PasswdUser

var InvalidFormat = errors.New("Invalid line format")

func parsePasswdLine(s string) (*PasswdUser, error) {
	ss := strings.Split(s, ":")
	if len(ss) <= 1 { return nil, InvalidFormat }
	uid, err := strconv.Atoi(ss[2])
	if err != nil { return nil, err }
	gid, err := strconv.Atoi(ss[3])
	if err != nil { return nil, err }
	return &PasswdUser{
		UserName: ss[0],
		Password: ss[1],
		UID: uid,
		GID: gid,
		Gecos: ss[4],
		HomeDir: ss[5],
		Shell: ss[6],
	}, nil
}

func ParsePasswdString(s string) (Passwd, error) {
	var res Passwd = make(map[string]*PasswdUser, 0)
	for item := range strings.SplitSeq(s, "\n") {
		if len(strings.TrimSpace(item)) <= 0 { continue }
		u, err := parsePasswdLine(item)
		if err != nil { return nil, err }
		res[u.UserName] = u
	}
	return res, nil
}

func (u *PasswdUser) String() string {
	return fmt.Sprintf(
		"%s:%s:%d:%d:%s:%s:%s",
		u.UserName,
		u.Password,
		u.UID,
		u.GID,
		u.Gecos,
		u.HomeDir,
		u.Shell,
	)
}

func (u Passwd) String() string {
	res := make([]string, 0)
	for _, item := range u {
		res = append(res, item.String())
	}
	return strings.Join(res, "\n")
}

func HasUser(username string) (bool, error) {
	s, err := os.ReadFile("/etc/passwd")
	if err != nil { return false, err }
	for item := range strings.SplitSeq(string(s), "\n") {
		if strings.HasPrefix(item, username + ":") { return true, nil }
	}
	return false, nil
}

func LoadPasswdFile() (Passwd, error) {
	s, err := os.ReadFile("/etc/passwd")
	if err != nil { return nil, err }
	return ParsePasswdString(string(s))
}

func GetUser(username string) (*PasswdUser, error) {
	pwd, err := LoadPasswdFile()
	if err != nil { return nil, err }
	u, ok := pwd[username]
	if !ok { return nil, nil }
	return u, nil
}

