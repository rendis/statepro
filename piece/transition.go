package piece

type GTransition[T any] struct {
	Guards []*GGuard[T] // At least one guard is required
}

func (t *GTransition[T]) resolve(c T, e Event) (*string, error) {
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
