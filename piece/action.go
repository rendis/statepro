package piece

type TAction[CTX any] func(CTX, GEvent)

type GAction[CTX any] struct {
	Name string
	Act  *TAction[CTX]
}

func (a *GAction[CTX]) execute(c CTX, e GEvent) {
	go (*a.Act)(c, e)
}
