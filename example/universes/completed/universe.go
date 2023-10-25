package completed

func NewAdmissionCompletedUniverse() *AdmissionCompletedUniverse {
	return &AdmissionCompletedUniverse{}
}

type AdmissionCompletedUniverse struct{}

func (a *AdmissionCompletedUniverse) GetUniverseId() string {
	return "admission-default-enrollment-completed"
}
