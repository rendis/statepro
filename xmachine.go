package statepro

// XMachine is the json representation of a machine.
// Contains Id, Initial state and the States, these fields are part of the XState.
// The SuccessFlow and Version fields are not part of the XState, but are used to define extra information about the machine.
// Version is Required, SuccessFlow is Optional.
type XMachine struct {
	Id          *string            `json:"id" bson:"id" xml:"id" yaml:"id"`
	Initial     *string            `json:"initial" bson:"initial" xml:"initial" yaml:"initial"`
	States      map[string]*XState `json:"states" bson:"states" xml:"states" yaml:"states"`
	SuccessFlow []string           `json:"successFlow" bson:"successFlow" xml:"successFlow" yaml:"successFlow"`
	Version     string             `json:"version" bson:"version" xml:"version" yaml:"version"`
}

// XState is the json representation of a state.
type XState struct {
	Always      []*XTransition            `json:"always" bson:"always" xml:"always" yaml:"always"`
	On          map[string][]*XTransition `json:"on" bson:"on" xml:"on" yaml:"on"`
	After       map[int][]*XTransition    `json:"after" bson:"after" xml:"after" yaml:"after"`
	Type        *string                   `json:"type" bson:"type" xml:"type" yaml:"type"`
	Invoke      []*XInvoke                `json:"invoke" bson:"invoke" xml:"invoke" yaml:"invoke"`
	Entry       []string                  `json:"entry" bson:"entry" xml:"entry" yaml:"entry"`
	Exit        []string                  `json:"exit" bson:"exit" xml:"exit" yaml:"exit"`
	Description *string                   `json:"description" bson:"description" xml:"description" yaml:"description"`
}

// XTransition is the json representation of a transition.
type XTransition struct {
	Condition   *string  `json:"cond" bson:"condition" xml:"condition" yaml:"condition"`
	Target      *string  `json:"target" bson:"target" xml:"target" yaml:"target"`
	Actions     []string `json:"actions" bson:"actions" xml:"actions" yaml:"actions"`
	Description *string  `json:"description" bson:"description" xml:"description" yaml:"description"`
}

// XInvoke is the json representation of the invoke property.
type XInvoke struct {
	Id      *string        `json:"id" bson:"id" xml:"id" yaml:"id"`
	Src     *string        `json:"src" bson:"src" xml:"src" yaml:"src"`
	OnDone  []*XTransition `json:"onDone" bson:"onDone" xml:"onDone" yaml:"onDone"`
	OnError []*XTransition `json:"onError" bson:"onError" xml:"onError" yaml:"onError"`
}
