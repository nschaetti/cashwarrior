package parser

import "errors"

func GetArgKindCount(parsedArgs ParsedCmdLine, filter bool) map[ArgKind]int {
	count := make(map[ArgKind]int)
	if !filter {
		for _, parsedToken := range parsedArgs.Args {
			count[parsedToken.ArgKind()]++
		}
		return count
	}
	for _, parsedToken := range parsedArgs.Filters {
		count[parsedToken.ArgKind()]++
	}
	return count
}

//func GetTokenKindCount(parsedTokens ParsedCmdLine, filter bool) map[TokenKind]int {
//	count := make(map[TokenKind]int)
//	if !filter {
//		for _, parsedToken := range parsedTokens.Args {
//			count[parsedToken.Kind]++
//		}
//		return count
//	}
//	for _, parsedToken := range parsedTokens.Filters {
//		count[parsedToken.Kind]++
//	}
//	return count
//}

// ParsedCmdLine contains the parsed command, filters, and command arguments.
type ParsedCmdLine struct {
	Command    string
	Subcommand string
	Filters    []Arg
	Args       []Arg
	Flags      []Arg
}

func (p *ParsedCmdLine) HasFlag(name string) bool {
	for _, arg := range p.Flags {
		flag, ok := arg.(ArgFlag)
		if !ok {
			continue
		}
		if flag.Key == name {
			return true
		}
	}
	return false
}

// GetFlagString returns the value of a string flag.
func (p *ParsedCmdLine) GetFlagString(name string) (string, bool) {
	for _, arg := range p.Flags {
		flag, ok := arg.(ArgFlag)
		if !ok || flag.Key != name {
			continue
		}
		value, ok := flag.Value.(StringItem)
		if !ok {
			return "", false
		}
		return value.Value, true
	}
	return "", false
}

func (p *ParsedCmdLine) GetArgAttributeValue(name string) (AttributeValue, error) {
	for _, arg := range p.Args {
		attr, ok := arg.(ArgAttribute)
		if !ok {
			continue
		}
		if attr.Key == name {
			return attr.Value, nil
		}
	}
	return AttributeValue{}, errors.New("argument not found")
}

func (p *ParsedCmdLine) Left() []Arg {
	return p.Filters
}

func (p *ParsedCmdLine) Right() []Arg {
	return p.Args
}

func (p *ParsedCmdLine) GetTokenKindCount(filter bool) map[ArgKind]int {
	return GetArgKindCount(*p, filter)
}

func (p *ParsedCmdLine) GetAttributesCount(filter bool) map[string]int {
	count := make(map[string]int)
	if !filter {
		for _, parsedArg := range p.Args {
			if parsedArg.ArgKind() != ArgKindAttribute {
				continue
			}
			attr, ok := parsedArg.(ArgAttribute)
			if !ok {
				continue
			}
			count[attr.Key]++
		}
		return count
	}
	for _, parsedArg := range p.Filters {
		if parsedArg.ArgKind() != ArgKindAttribute {
			continue
		}
		attr, ok := parsedArg.(ArgAttribute)
		if !ok {
			continue
		}
		count[attr.Key]++
	}
	return count
}

func (p *ParsedCmdLine) RemoveByKind(kind ArgKind) {
	dst := make([]Arg, 0, len(p.Args))
	for _, token := range p.Args {
		if token.ArgKind() != kind {
			dst = append(dst, token)
		}
	}
	p.Args = dst
}

func (p *ParsedCmdLine) Append(a Arg, filter bool) {
	if filter {
		p.Filters = append(p.Filters, a)
		return
	}
	p.Args = append(p.Args, a)
}
