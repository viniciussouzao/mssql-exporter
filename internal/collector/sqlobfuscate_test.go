package collector

import "testing"

func TestObfuscateSQL(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"string literal", "SELECT * FROM t WHERE cpf = '123.456.789-00'", "SELECT * FROM t WHERE cpf = ?"},
		{"escaped quote", "SELECT 1 FROM t WHERE name = 'O''Brien' AND x = 1", "SELECT ? FROM t WHERE name = ? AND x = ?"},
		{"unicode string", "UPDATE t SET note = N'ação' WHERE id = 7", "UPDATE t SET note = ? WHERE id = ?"},
		{"N as identifier end is not a prefix", "SELECT colN FROM t WHERE colN='x'", "SELECT colN FROM t WHERE colN=?"},
		{"numbers", "SELECT TOP (50) a FROM t WHERE v > 3.14 AND w < -2 AND e = 1.5e10 AND f = .5", "SELECT TOP (?) a FROM t WHERE v > ? AND w < -? AND e = ? AND f = ?"},
		{"hex", "SELECT * FROM t WHERE h = 0x1A2B", "SELECT * FROM t WHERE h = ?"},
		{"digits inside identifiers kept", "SELECT t1.col2, @p1, #tmp3 FROM dbo.Table9 t1", "SELECT t1.col2, @p1, #tmp3 FROM dbo.Table9 t1"},
		{"delimited identifiers kept", `SELECT [Order 1], "Col 2" FROM [dbo].[T]]x]`, `SELECT [Order 1], "Col 2" FROM [dbo].[T]]x]`},
		{"bracket digits not obfuscated after", "SELECT [a]1", "SELECT [a]1"},
		{"line comment removed", "SELECT a -- secret 123\nFROM t", "SELECT a FROM t"},
		{"nested block comment removed", "SELECT /* outer /* inner 'x' */ still */ a FROM t", "SELECT a FROM t"},
		{"whitespace collapsed", "SELECT\n\ta,\r\n  b\tFROM   t", "SELECT a, b FROM t"},
		{"unterminated string", "SELECT 'abc", "SELECT ?"},
		{"IN list", "WHERE id IN (1, 2, 3)", "WHERE id IN (?, ?, ?)"},
		{"waitfor", "WAITFOR DELAY '00:00:30'", "WAITFOR DELAY ?"},
		{"money", "WHERE salary = $12345.67 OR fee < $5 OR c = -$45.56", "WHERE salary = ? OR fee < ? OR c = -?"},
		{"signed money", "SELECT $-1.25, $+3", "SELECT ?, ?"},
		{"money without integer part", "SELECT $.50", "SELECT ?"},
		{"$ pseudo-columns kept", "SELECT $action, $partition.pf(c), a$1 FROM t", "SELECT $action, $partition.pf(c), a$1 FROM t"},
		{"other currency symbols", "SELECT €50, £3.5", "SELECT €?, £?"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := obfuscateSQL(tc.in, 0, true); got != tc.want {
				t.Errorf("obfuscateSQL(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestObfuscateSQLTruncates(t *testing.T) {
	got := obfuscateSQL("SELECT a, b, c FROM some_table WHERE x = 'secret'", 20, true)
	if want := "SELECT a, b, c FR..."; got != want {
		t.Errorf("obfuscateSQL(..., 20) = %q, want %q", got, want)
	}
	// Runes, not bytes: multi-byte characters are never split.
	got = obfuscateSQL("SELECT ação_çã FROM t", 12, true)
	if want := "SELECT aç..."; got != want {
		t.Errorf("obfuscateSQL(multibyte, 12) = %q, want %q", got, want)
	}
	if got := obfuscateSQL("SELECT 1", 120, true); got != "SELECT ?" {
		t.Errorf("obfuscateSQL(short) = %q, want no truncation", got)
	}
}

func TestObfuscateSQLInvalidUTF8(t *testing.T) {
	got := obfuscateSQL("SELECT a\xff FROM t", 0, true)
	if got != "SELECT a FROM t" {
		t.Errorf("obfuscateSQL(invalid UTF-8) = %q, want invalid bytes dropped", got)
	}
}

func TestObfuscateSQLQuotedIdentifierOff(t *testing.T) {
	// With QUOTED_IDENTIFIER OFF "..." is a string literal.
	in := `SELECT "Col 2" FROM t WHERE email = "alice@example.invalid" AND n = "say ""hi"""`
	if got, want := obfuscateSQL(in, 0, false), "SELECT ? FROM t WHERE email = ? AND n = ?"; got != want {
		t.Errorf("obfuscateSQL(quoted identifier off) = %q, want %q", got, want)
	}
	// Brackets are identifiers either way.
	if got, want := obfuscateSQL("SELECT [a b] FROM t", 0, false), "SELECT [a b] FROM t"; got != want {
		t.Errorf("obfuscateSQL(brackets, quoted identifier off) = %q, want %q", got, want)
	}
}
