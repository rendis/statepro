package waiting_confirmation

func NewAdmissionWaitingConfirmationUniverse() *AdmissionWaitingConfirmationUniverse {
	return &AdmissionWaitingConfirmationUniverse{}
}

type AdmissionWaitingConfirmationUniverse struct{}

func (a *AdmissionWaitingConfirmationUniverse) GetUniverseId() string {
	return "admission_default_waiting_confirmation_universe"
}
