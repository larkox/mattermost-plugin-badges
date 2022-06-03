package commandparser

import "unicode"

const (
	stateLooking = "looking"
	stateQuoted  = "quoted"
	stateEscape  = "escape"
	stateFilling = "filling"
)

func Parse(in string) []string {
	nextArg := ""
	out := []string{}
	state := stateLooking
	for _, c := range in {
		switch state {
		case stateLooking:
			if unicode.IsSpace(c) {
				continue
			}
			if c == '"' {
				state = stateQuoted
				continue
			}

			state = stateFilling
			nextArg += string(c)
		case stateFilling:
			if unicode.IsSpace(c) {
				state = stateLooking
				out = append(out, nextArg)
				nextArg = ""
				continue
			}

			nextArg += string(c)
		case stateQuoted:
			if c == '"' {
				state = stateLooking
				out = append(out, nextArg)
				nextArg = ""
				continue
			}

			if c == '\\' {
				state = stateEscape
				continue
			}

			nextArg += string(c)
		case stateEscape:
			if c == '"' {
				state = stateQuoted
				nextArg += string(c)
				continue
			}

			// Not escaping quotes, so we readd the backslash
			nextArg += string('\\')
			nextArg += string(c)
		default:
			panic("unexpected state")
		}
	}

	if state == stateEscape {
		nextArg += string('\\')
	}

	if state == stateFilling {
		out = append(out, nextArg)
	}

	return out
}
