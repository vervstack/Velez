//go:build ruleguard

package gorules

import "github.com/quasilyte/go-ruleguard/dsl"

func bareErrorReturn(m dsl.Matcher) {
	m.Match(`return $*_, $err`, `return $err`).
		Where(m["err"].Type.Implements("error") && (m["err"].Node.Is("Ident") || m["err"].Node.Is("SelectorExpr")) && m["err"].Text != "nil").
		Report("returned error must be wrapped via rerrors.Wrap")
}
