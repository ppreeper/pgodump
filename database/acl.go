package database

import (
	"context"
	"fmt"
	"log"
	"strings"
)

func (db *Database) GetOwnershipAndPrivileges(ctx context.Context, schema, name, kind string) (string, error) {
	ctx, cancel := db.withTimeout(ctx)
	defer cancel()

	q := `
		SELECT
			pg_get_userbyid(c.relowner) AS owner,
			c.relacl AS acl
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2;`

	type RelInfo struct {
		Owner string  `db:"owner"`
		ACL   *string `db:"acl"`
	}

	var info RelInfo
	if err := db.GetContext(ctx, &info, q, schema, name); err != nil {
		return "", fmt.Errorf("get ACL for %s.%s: %w", schema, name, err)
	}

	fullName := QuoteIdentifier(schema) + "." + QuoteIdentifier(name)

	var b strings.Builder

	// 1. Ownership
	if !db.NoOwner {
		fmt.Fprintf(&b, "ALTER %s %s OWNER TO %s;\n", kind, fullName, QuoteIdentifier(info.Owner))
	}

	// 2. Privileges — only emit when explicit ACL entries exist.
	// A nil ACL means the object inherits default privileges; emitting a
	// blanket REVOKE ALL without corresponding GRANTs would silently remove
	// access that was never explicitly set.
	if !db.NoPrivileges && info.ACL != nil {
		fmt.Fprintf(&b, "REVOKE ALL ON %s %s FROM PUBLIC;\n", kind, fullName)

		aclStr := strings.Trim(*info.ACL, "{}")
		if aclStr != "" {
			entries := strings.Split(aclStr, ",")
			for _, entry := range entries {
				eqIdx := strings.IndexByte(entry, '=')
				if eqIdx < 0 {
					return "", fmt.Errorf("malformed ACL entry %q for %s.%s", entry, schema, name)
				}
				granteeRaw := entry[:eqIdx]
				rest := entry[eqIdx+1:]
				var grantee string
				if granteeRaw == "" {
					grantee = "PUBLIC"
				} else {
					if granteeRaw == info.Owner {
						continue
					}
					grantee = QuoteIdentifier(granteeRaw)
				}

				privsParts := strings.Split(rest, "/")
				privsCode := privsParts[0]

				var grants []string
				for _, code := range privsCode {
					switch code {
					case 'a':
						grants = append(grants, "INSERT")
					case 'r':
						grants = append(grants, "SELECT")
					case 'w':
						grants = append(grants, "UPDATE")
					case 'd':
						grants = append(grants, "DELETE")
					case 'D':
						grants = append(grants, "TRUNCATE")
					case 'x':
						grants = append(grants, "REFERENCES")
					case 't':
						grants = append(grants, "TRIGGER")
					case 'm':
						grants = append(grants, "MAINTAIN")
					case 'U':
						grants = append(grants, "USAGE")
					case 'C':
						grants = append(grants, "CREATE")
					case 'X':
						grants = append(grants, "EXECUTE")
					default:
						log.Printf("WARNING: unknown ACL privilege code %q for %s.%s", code, schema, name)
					}
				}

				if len(grants) > 0 {
					fmt.Fprintf(&b, "GRANT %s ON %s %s TO %s;\n", strings.Join(grants, ", "), kind, fullName, grantee)
				}
			}
		}
	}

	result := b.String()
	if result == "" {
		return "", nil
	}
	return result + "\n", nil
}
