package piece

type GTransition[ContextType any] struct {
	Guards []*GGuard[ContextType] // At least one guard is required
}

func (t *GTransition[ContextType]) resolve(c ContextType, e Event) (*string, error) {
	for _, g := range t.Guards {
		target, ok, err := g.check(c, e)
		if err != nil {
			return nil, err
		}
		if ok {
			return &target, nil
		}
	}
	return nil, nil
}
