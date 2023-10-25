package payment

func NewPaymentUniverse() *PaymentUniverse {
	return &PaymentUniverse{}
}

type PaymentUniverse struct{}

func (a *PaymentUniverse) GetUniverseId() string {
	return "admission_default_payment_universe"
}
