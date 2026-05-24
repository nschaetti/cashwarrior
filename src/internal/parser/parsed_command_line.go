package parser

func GetTokenKindCount(parsedTokens ParsedCmdLine, filter bool) map[TokenKind]int {
	count := make(map[TokenKind]int)
	if !filter {
		for _, parsedToken := range parsedTokens.Args {
			count[parsedToken.Kind]++
		}
		return count
	}
	for _, parsedToken := range parsedTokens.Filters {
		count[parsedToken.Kind]++
	}
	return count
}

// ParsedCmdLine contains the parsed command, filters, and command arguments.
type ParsedCmdLine struct {
	Command string
	Filters []Token
	Args    []Token
}

func (p *ParsedCmdLine) GetTokenKindCount(filter bool) map[TokenKind]int {
	return GetTokenKindCount(*p, filter)
}

func (p *ParsedCmdLine) GetAmount() []*Token {
	var amounts []*Token
	for i := range p.Args {
		if p.Args[i].Kind == TokenAmount {
			amounts = append(amounts, &p.Args[i])
		}
	}
	return amounts
}

func (p *ParsedCmdLine) RemoveByKind(kind TokenKind) {
	dst := p.Args[:0]
	for _, token := range p.Args {
		if token.Kind != kind {
			dst = append(dst, token)
		}
	}
	p.Args = dst
}

func (p *ParsedCmdLine) Append(t Token, filter bool) {
	if filter {
		p.Filters = append(p.Filters, t)
		return
	}
	p.Args = append(p.Args, t)
}
