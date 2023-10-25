package form

func NewFormUniverse() *FormUniverse {
	return &FormUniverse{}
}

type FormUniverse struct{}

func (a *FormUniverse) GetUniverseId() string {
	return "admission_default_form_universe"
}
