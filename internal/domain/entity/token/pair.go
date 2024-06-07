package token

import "fmt"

type Pair struct {
	base  *Token
	quote *Token
}

func (p *Pair) BaseToken() *Token  { return p.base }
func (p *Pair) QuoteToken() *Token { return p.quote }
func (p *Pair) String() string     { return fmt.Sprintf("%s/%s", p.base, p.quote) }

func NewPair(base, quote *Token) *Pair {
	return &Pair{base: base, quote: quote}
}
