package ssh

import (
	"fmt"
	"maps"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/passwd"
)

type SSHKeyManagingContext struct {
	keyFilePath string
	Managed map[string]map[string]string
	notManaged []string
}

func parseKey(s string) (bool, string, string, string, error) {
	// we expect all keys managed by aegis should have the prefix
	//     command="aegis ssh {username} {keyname}"
	var r, err = regexp.Compile("\\s*command=\"aegis ssh ([^ ]*) ([^ ]*)\"\\s*(.*)")	
	if err != nil { return false, "", "", "", err }
	k := r.FindSubmatch([]byte(s))
	if len(k) <= 0 { return false, "", "", "", nil }
	return true, string(k[1]), string(k[2]), string(k[3]), nil
}

func ToContext(cfg *aegis.AegisConfig) (*SSHKeyManagingContext, error) {
	u, err := passwd.GetUser(cfg.GitUser)
	if err != nil { return nil, err }
	p := path.Join(u.HomeDir, ".ssh", "authorized_keys")
	f, err := os.ReadFile(p)
	if err != nil { return nil, err }
	managed := make(map[string]map[string]string, 0)
	currentName := ""
	currentManaged := make(map[string]string, 0)
	notManaged := make([]string, 0)
	for k := range strings.SplitSeq(string(f), "\n") {
		kstr := strings.TrimSpace(k)
		if len(kstr) <= 0 { continue }
		if strings.HasPrefix(kstr, "#") { continue }
		chk, userName, keyName, key, err := parseKey(k)
		if err != nil { return nil, err }
		if !chk { notManaged = append(notManaged, k); continue }
		if currentName == "" { currentName = userName }
		if userName != currentName {
			managed[currentName] = currentManaged
			d, ok := managed[userName]
			if ok {
				currentManaged = d
			} else {
				currentManaged = make(map[string]string, 0)
			}
			currentManaged[keyName] = key
		} else {
			currentManaged[keyName] = key
		}
	}
	_, ok := managed[currentName]
	if !ok {
		managed[currentName] = currentManaged
	} else {
		maps.Copy(managed[currentName], currentManaged)
	}
	return &SSHKeyManagingContext{
		keyFilePath: p,
		Managed: managed,
		notManaged: notManaged,
	}, nil
}

func (ctx *SSHKeyManagingContext) Sync() error {
	f, err := os.OpenFile(ctx.keyFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil { return err }
	for userName, pack := range ctx.Managed {
		for keyName, key := range pack {
			_, err := f.WriteString(fmt.Sprintf("command=\"aegis ssh %s %s\" %s", userName, keyName, key))
			if err != nil { return err }
			_, err = f.WriteString("\n")
			if err != nil { return err }
		}
		f.WriteString("\n")
	}
	_, err = f.WriteString("\n")
	if err != nil { return err }
	for _, item := range ctx.notManaged {
		_, err := f.WriteString(item)
		if err != nil { return err }
		_, err = f.WriteString("\n")
		if err != nil { return err }
	}
	return nil
}

func (ctx *SSHKeyManagingContext) AddAuthorizedKey(username string, keyname string, key string) {
	pack, ok := ctx.Managed[username]
	if !ok {
		pack = make(map[string]string, 0)
		ctx.Managed[username] = pack
	}
	pack[keyname] = key
}

func (ctx *SSHKeyManagingContext) RemoveAuthorizedKey(username string, keyname string) {
	pack, ok := ctx.Managed[username]
	if !ok { return }
	delete(pack, keyname)
}

func (ctx *SSHKeyManagingContext) GetAuthorizedKey(username string, keyname string) string {
	pack, ok := ctx.Managed[username]
	if !ok { return "" }
	s, ok := pack[keyname]
	if !ok { return "" }
	return s
}
