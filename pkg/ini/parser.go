package ini

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strings"
)

// a simple-as-fvck ini parser, made to parse git config files.
// syntax from:
//     https://git-scm.com/docs/git-config#_configuration_file
// should this be a part of gitlib? i don't think so - some people
// might like to use git ini on its own.

// ID = [a-zA-Z0-9-.]+
// SectionHeader := "[" ID ("." ID | WS "\"" Subsection "\"")? "]"
// Subsection = ((?:[^"\\]|\\"|\\\\)*)
// KVPair := ID (WS "=" WS Value)?
// Value := QuotedValue | UnquotedValue
// UnquotedValue = [^#;\n](?:(?[^#;\n]|\\\n))*[^#;\n])?
// QuotedValue = "(?[^"\\]|\\"|\\\n|\\\\)*"
// Comment = [#;][^\n]*
// Section := SectionHeader (KVPair | Comment)*
// Document := Section*

// this, of course, does not strictly follow the rules above

func processQuotedValueString(s string) string {
	res := strings.Builder{}
	ss := []rune(s)
	i := 0
	for i < len(ss) {
		if ss[i] == '\\' {
			if i+1 >= len(ss) { break }
			switch ss[i+1] {
			case 'n':
				res.WriteRune('\n')
			case 't':
				res.WriteRune('\t')
			case 'b':
				res.WriteRune('\b')
			default:
				res.WriteRune(ss[i+1])
			}
			i += 2
		} else {
			res.WriteRune(ss[i])
			i += 1
		}
	}
	return res.String()
}

var reQuotedContinuedLine = regexp.MustCompile("^\\s*(.*?)((?:\\\\\\s*|\"\\s*(?:[#;].*)))$")

func parseQuotedContinuedLine(line string) (string, bool, error) {
	// [content] [\] ====> [content] (continued)
	// [content] ["] [comment?] ====> [content] (ended)
	// * ====> syntax error
	re := reQuotedContinuedLine
	vs := re.FindStringSubmatch(line)
	if strings.HasPrefix(vs[2], "\\") { return processQuotedValueString(vs[1]), true, nil }
	if strings.HasPrefix(vs[2], "\"") { return processQuotedValueString(vs[1]), false, nil }
	return "", false, errors.New("Line failed to explicitly end or continue.")
}

var reContinuedLine = regexp.MustCompile(`^\s*(.*?)(\s*)(\\?)\s*(?:[#;].*)?$`)
func parseContinuedLine(line string) (string, bool, error) {
	//  [content] [comment?]  ====>  [content] (ended)
	//  [content] [ws] [comment?]  ====>  [content] (ended)
	//  [content] [\] [comment?]  ====>  [content] (ended)
	//  [content] [ws] [\] [comment?]  ====>  [content] [ws] (continued)
	re := reContinuedLine
	vs := re.FindStringSubmatch(line)
	continued := len(vs[3]) <= 0
	if len(vs[2]) <= 0 {
		if continued {
			// content, ws, \, comment?
			return vs[1] + vs[2], true, nil
		} else {
			// content, ws, comment?
			return vs[1], false, nil
		}
	} else {
		return vs[1], continued, nil
	}
}

var reProcessHeaderQuotedString = regexp.MustCompile(`\\(.)`)
func processHeaderQuotedString(s string) (string, error) {
	re := reProcessHeaderQuotedString
	return re.ReplaceAllString(s, "$1"), nil
}
var reHeader = regexp.MustCompile("^\\s*\\[\\s*([a-zA-Z0-9-.]+?)(?:\\.([a-zA-Z0-9-.]+)|\\s+\"((?:[^\\\"]|\\\\|\\\")*)\")?\\s*\\]\\s*(?:[#;].*)?$")
func parseHeader(line string) (string, string, error) {
	re := reHeader
	v := re.FindStringSubmatch(line)
	if len(v[2]) <= 0 {
		r, err := processHeaderQuotedString(v[3])
		if err != nil { return "", "", err }
		return v[1], r, nil
	} else {
		return v[1], v[2], nil
	}
}
var reKVPairLine = regexp.MustCompile(`^\s*([a-zA-Z0-9-.]+)\s*(?:(=)\s*(.*)|([#;].*))?\s*$`)
var reKVPairLine2 = regexp.MustCompile("^\\s*([^\"].*?)(?:(\\\\)\\s*|[#;].*)$")
var reKVPairLine3 = regexp.MustCompile("^\\s*(\\\")((?:[^\"\\\\]|\\\\\"|\\\\\\\\)*)(?:(\")(?:[#;].*)?|(\\\\)\\s*)")
func parseKVPairLine(line string) (string, string, bool, bool, error) {
	// return value: key, value (first line), quoted, continued, error
	// [key] [comment?] ==> [key], "true", unquoted, ended
	// [key] [=] [value] [comment?] ==> [key], [value], unquoted, ended
	// [key] [=] ["] [value] ["] [comment?] ==> [key], [value], quoted, ended
	// [key] [=] ["] [value] [\] ==> [key], [value], quoted, continued
	// [key] [=] ["] [value] *  ==> syntax error: quoted value
	// * ==> syntax error
	// 1.  check for [key] and [=]
	re := reKVPairLine
	v := re.FindStringSubmatch(line)
	if len(v[2]) <= 0 { return v[1], v[1], false, false, nil }
	key := v[1]
	subj := v[3]
	// 2.  check that if there's no ["]
	re2 := reKVPairLine2
	v = re2.FindStringSubmatch(subj)
	quoted := false
	continued := false
	if len(v) > 0 {
		if len(v[2]) > 0 { continued = true }
		return key, v[1], quoted, continued, nil
	}
	// 3. check for ["]
	re3 := reKVPairLine3
	v = re3.FindStringSubmatch(subj)
	if len(v) <= 0 {
		return key, strings.TrimSpace(subj), false, false, nil
	}
	if len(v[1]) <= 0 { return "", "", false, false, errors.New("Invalid syntax") }
	if len(v[3]) > 0 {
		// the value is ending within this line.
		return key, processQuotedValueString(v[2]), true, false, nil
	}
	if len(v[4]) > 0 {
		return key, processQuotedValueString(v[2]), true, true, nil
	}
	return "", "", false, false, errors.New("Invalid syntax")
}

func ParseINI(o io.Reader) (INI, error) {
	var err error
	res := make(INI, 0)
	sc := bufio.NewScanner(o)
	sec := ""
	subsec := ""
	secbuf := make(map[string]string)
	buf := make([]string, 0)
	key := ""
	quoted := false
	continued := false
	// `quoted` and `continued` only triggers when multiline escape
	// happens; if there's no multiline escape it would've simply
	// be done within the parse of the line.
	for sc.Scan() {
		line := strings.TrimLeft(sc.Text(), " \t")
		if continued {
			var l string
			if quoted {
				l, continued, err = parseQuotedContinuedLine(line)
			} else {
				l, continued, err = parseContinuedLine(line)
			}
			if err != nil { return nil, err }
			buf = append(buf, l)
			if !continued {
				quoted = false
				secbuf[key] = strings.Join(buf, "")
				buf = make([]string, 0)
			}
			continue
		}
		if strings.HasPrefix(line, "#") { continue }
		if strings.HasPrefix(line, ";") { continue }
		if strings.HasPrefix(line, "[") {
			// header.
			res.InsertConfigBatch(sec, subsec, secbuf)
			secbuf = make(map[string]string)
			sec, subsec, err = parseHeader(line)
			if err != nil { return nil, err }
			continue
		}
		// kvpair.
		var v string
		key, v, quoted, continued, err = parseKVPairLine(line)
		if continued {
			buf = append(buf, v)
		} else {
			secbuf[key] = v
			key = ""
		}
	}
	if len(buf) > 0 { return nil, errors.New("Unfinished quoted string") }
	if len(secbuf) > 0 {
		res.InsertConfigBatch(sec, subsec, secbuf)
	}
	return res, nil
}


