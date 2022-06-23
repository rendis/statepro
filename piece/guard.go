package piece

type TPredicate[CTX any] func(CTX, GEvent) bool

type GGuard[CTX any] struct {
	Condition *string
	Pred      *TPredicate[CTX]
	Target    *string
	Actions   []*GAction[CTX]
}

func (g *GGuard[CTX]) check(c CTX, e GEvent) (bool, string) {
	// (Else || Directly) GGuard
	if g.Condition == nil || g.Pred == nil {
		g.execute(c, e)
		return true, *g.Target
	}
	// (If || ElseIf) GGuard
	if (*g.Pred)(c, e) {
		g.execute(c, e)
		return true, *g.Target
	}
	return false, ""
}

func (g *GGuard[CTX]) execute(c CTX, e GEvent) {
	for _, a := range g.Actions {
		a.execute(c, e)
	}
}
