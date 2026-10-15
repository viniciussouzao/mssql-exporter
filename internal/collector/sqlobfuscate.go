package collector

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// obfuscateSQL replaces the literals of a T-SQL statement with `?` and
// removes its comments, so statement text can be exposed as a label without
// leaking the values ad hoc queries carry (customer data, IDs). Whitespace
// runs collapse to one space and the result is cut to maxLen characters
// ("..." marks a cut). Identifiers, including [bracketed] and "quoted"
// ones, and @variables are kept.
func obfuscateSQL(sql string, maxLen int) string {
	sql = strings.ToValidUTF8(sql, "")
	var b strings.Builder
	b.Grow(len(sql))

	// identChar is true when the previous rune, with no space or comment in
	// between, belongs to a word (identifier, keyword, variable): a digit
	// there is part of the name, not a numeric literal.
	identChar := false
	space := false
	emit := func(s string) {
		if space && b.Len() > 0 {
			b.WriteByte(' ')
		}
		space = false
		b.WriteString(s)
		r, _ := utf8.DecodeLastRuneInString(s)
		identChar = isIdentRune(r)
	}

	for i := 0; i < len(sql); {
		r, size := utf8.DecodeRuneInString(sql[i:])
		switch {
		case unicode.IsSpace(r):
			space, identChar = true, false
			i += size

		case strings.HasPrefix(sql[i:], "--"):
			end := strings.IndexByte(sql[i:], '\n')
			if end < 0 {
				end = len(sql) - i
			}
			space, identChar = true, false
			i += end

		case strings.HasPrefix(sql[i:], "/*"):
			// T-SQL block comments nest.
			depth, j := 0, i
			for j < len(sql) {
				if strings.HasPrefix(sql[j:], "/*") {
					depth++
					j += 2
				} else if strings.HasPrefix(sql[j:], "*/") {
					depth--
					j += 2
					if depth == 0 {
						break
					}
				} else {
					j++
				}
			}
			space, identChar = true, false
			i = j

		case r == '\'' || ((r == 'N' || r == 'n') && !identChar && strings.HasPrefix(sql[i+1:], "'")):
			// 'string' and N'unicode string'; '' is an escaped quote.
			j := i + strings.IndexByte(sql[i:], '\'') + 1
			for j < len(sql) {
				if sql[j] == '\'' {
					if j+1 < len(sql) && sql[j+1] == '\'' {
						j += 2
						continue
					}
					j++
					break
				}
				j++
			}
			emit("?")
			i = j

		case r == '[' || r == '"':
			// Delimited identifier, copied as is; ]] and "" are escapes.
			closing := byte(']')
			if r == '"' {
				closing = '"'
			}
			j := i + 1
			for j < len(sql) {
				if sql[j] == closing {
					if j+1 < len(sql) && sql[j+1] == closing {
						j += 2
						continue
					}
					j++
					break
				}
				j++
			}
			emit(sql[i:j])
			identChar = true
			i = j

		case !identChar && (isDigit(sql[i]) || (sql[i] == '.' && i+1 < len(sql) && isDigit(sql[i+1]))):
			j := i
			if strings.HasPrefix(sql[i:], "0x") || strings.HasPrefix(sql[i:], "0X") {
				j += 2
				for j < len(sql) && isHexDigit(sql[j]) {
					j++
				}
			} else {
				for j < len(sql) && (isDigit(sql[j]) || sql[j] == '.') {
					j++
				}
				// Exponent: 1.5e10, 2E-3.
				if j < len(sql) && (sql[j] == 'e' || sql[j] == 'E') {
					k := j + 1
					if k < len(sql) && (sql[k] == '+' || sql[k] == '-') {
						k++
					}
					if k < len(sql) && isDigit(sql[k]) {
						for k < len(sql) && isDigit(sql[k]) {
							k++
						}
						j = k
					}
				}
			}
			emit("?")
			identChar = false
			i = j

		default:
			emit(sql[i : i+size])
			i += size
		}
	}

	return truncateRunes(b.String(), maxLen)
}

func truncateRunes(s string, maxLen int) string {
	if maxLen <= 0 || utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	const marker = "..."
	keep := max(maxLen-len(marker), 0)
	n := 0
	for i := range s {
		if n == keep {
			return s[:i] + marker
		}
		n++
	}
	return s
}

func isIdentRune(r rune) bool {
	return r == '_' || r == '@' || r == '#' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isHexDigit(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
