package hydrate

type Segment interface {
	Render() []byte
}

type LiteralSegment struct {
	data []byte
}

func (s *LiteralSegment) Render() []byte {
	return s.data
}

type PlaceholderSegment struct {
	id []byte
}

func (s *PlaceholderSegment) Render() []byte {
	return s.id
}
