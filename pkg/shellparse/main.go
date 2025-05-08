package shellparse

import "strings"

// this is a simple package that parses a simpler version of shell command,
// which should be enough to handle cmd issued by git clients.
// see also: https://github.com/git/git/blob/master/quote.c#L28
//
//     /* Help to copy the thing properly quoted for the shell safety.
//      * any single quote is replaced with '\'', any exclamation point
//      * is replaced with '\!', and the whole thing is enclosed in a
//      * single quote pair.
//      *
//      * E.g.
//      *  original     sq_quote     result
//      *  name     ==> name      ==> 'name'
//      *  a b      ==> a b       ==> 'a b'
//      *  a'b      ==> a'\''b    ==> 'a'\''b'
//      *  a!b      ==> a'\!'b    ==> 'a'\!'b'
//      */

// everything about this is not compatible with utf-8 but since this
// is used for dealing with urls we *should* be fine.

func isWhiteSpace(r byte) bool {
	return (r == byte(' ') || r == byte('\t') || r == byte('\b'))
}

func ParseShellCommand(s string) []string {
	res := make([]string, 0)
	buf := make([]string, 0)

	lens := len(s)
	i := 0
	for i < lens {
		if isWhiteSpace(s[i]) {
			if len(buf) > 0 {
				res = append(res, strings.Join(buf, ""))
				buf = make([]string, 0)
			}
			for isWhiteSpace(s[i]) && i < lens { i += 1 }
		}
		if s[i] == '\'' {
			i += 1
			buf = make([]string, 0)
			for i < lens && s[i] != '\''{
				if i + 4 < lens {
					k := s[i:i+4]
					if k == "'\\''" {
						buf = append(buf, k)
						i += 4
						continue
					} else if k == "'\\!'" {
						buf = append(buf, k)
						i += 4
						continue
					}
				}
				buf = append(buf, s[i:i+1])
				i += 1
			}
			if i < lens && s[i] == '\'' { i += 1 }
			res = append(res, strings.Join(buf, ""))
		} else {
			st := i
			for i < lens && !isWhiteSpace(s[i]) { i += 1 }
			res = append(res, s[st:i])
		}
	}
	return res
}

func Quote(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "'", "'\\''"), "!", "'\\!'")
}

