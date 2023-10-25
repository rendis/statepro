package form

func NewFormUniverse() *FormUniverse {
	return &FormUniverse{}
}

type FormUniverse struct{}

func (a *FormUniverse) GetUniverseId() string {
	return "admission-default-form-universe"
}
