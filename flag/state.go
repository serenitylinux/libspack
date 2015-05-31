package flag

type State int

const (
	Enabled  State = iota // +
	Disabled              // -
	Inherit               // ?
	Invert                // ~
	Invalid  State = -1
)

func StateFromString(s string) State {
	switch s {
	case "+":
		return Enabled
	case "-":
		return Disabled
	case "?":
		return Inherit
	case "~":
		return Invert
	default:
		return Invalid
	}
}
func StateFromBool(b bool) State {
	if b {
		return Enabled
	} else {
		return Disabled
	}
}

func (s State) String() string {
	switch s {
	case Enabled:
		return "+"
	case Disabled:
		return "-"
	case Inherit:
		return "?"
	case Invert:
		return "~"
	default:
		panic("!!!Invalid State!!!")
	}
}
