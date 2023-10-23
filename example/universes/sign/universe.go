package sign

func NewSignUniverse() *SignUniverse {
	return &SignUniverse{}
}

type SignUniverse struct{}

func (a *SignUniverse) GetUniverseId() string {
	return "admission_default_sign_universe"
}

func (a *SignUniverse) GetUniverseDescription() string {
	return "default universe for sign handling"
}