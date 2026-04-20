package database

import "testing"

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple", "users", `"users"`},
		{"reserved word", "select", `"select"`},
		{"with space", "my table", `"my table"`},
		{"embedded double quote", `foo"bar`, `"foo""bar"`},
		{"multiple double quotes", `a"b"c`, `"a""b""c"`},
		{"dot-qualified", "public.users", `"public"."users"`},
		{"dot-qualified with quote", `my"schema.my"table`, `"my""schema"."my""table"`},
		{"already quoted", `"users"`, `"users"`},
		{"empty string", "", `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuoteIdentifier(tt.in)
			if got != tt.want {
				t.Errorf("QuoteIdentifier(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestQuoteConstraint(t *testing.T) {
	tests := []struct {
		name    string
		def     string
		contype string
		want    string
	}{
		{
			"non-FK passthrough",
			"PRIMARY KEY (id)",
			"p",
			"PRIMARY KEY (id)",
		},
		{
			"simple FK",
			"FOREIGN KEY (user_id) REFERENCES users(id)",
			"f",
			`FOREIGN KEY ("user_id") REFERENCES "users"("id")`,
		},
		{
			"multi-column FK",
			"FOREIGN KEY (a, b) REFERENCES other(x, y)",
			"f",
			`FOREIGN KEY ("a", "b") REFERENCES "other"("x", "y")`,
		},
		{
			"schema-qualified FK",
			"FOREIGN KEY (user_id) REFERENCES public.users(id)",
			"f",
			`FOREIGN KEY ("user_id") REFERENCES "public"."users"("id")`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuoteConstraint(tt.def, tt.contype)
			if got != tt.want {
				t.Errorf("QuoteConstraint(%q, %q) =\n  %q\nwant\n  %q", tt.def, tt.contype, got, tt.want)
			}
		})
	}
}

func TestQuoteIndex(t *testing.T) {
	tests := []struct {
		name string
		def  string
		want string
	}{
		{
			"simple btree",
			"CREATE INDEX idx_name ON public.users USING btree (name)",
			`CREATE INDEX "idx_name" ON "public"."users" USING btree (name)`,
		},
		{
			"unique index",
			"CREATE UNIQUE INDEX idx_email ON users USING btree (email)",
			`CREATE UNIQUE INDEX "idx_email" ON "users" USING btree (email)`,
		},
		{
			"no match passthrough",
			"some random string",
			"some random string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuoteIndex(tt.def)
			if got != tt.want {
				t.Errorf("QuoteIndex(%q) =\n  %q\nwant\n  %q", tt.def, got, tt.want)
			}
		})
	}
}
