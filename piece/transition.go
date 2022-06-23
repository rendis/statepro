package piece

type GTransition[CTX any] struct {
	Guards []*GGuard[CTX]
}

func (t *GTransition[CTX]) resolve(c CTX, e GEvent) string {
	for _, g := range t.Guards {
		if ok, target := g.check(c, e); ok {
			return target
		}
	}
	return ""
}
