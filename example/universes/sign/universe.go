package sign

func NewSignUniverse() *SignUniverse {
	return &SignUniverse{}
}

type SignUniverse struct{}

func (a *SignUniverse) GetUniverseId() string {
	return "admission-default-sign-universe"
}
