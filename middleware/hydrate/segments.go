package hydrate

type Segment interface {
	Render(hydrator Hydrator) []byte
}

type LiteralSegment struct {
	data []byte
}

func (s *LiteralSegment) Render(hydrator Hydrator) []byte {
	return s.data
}

type PlaceholderSegment struct {
	id []byte
}

func (s *PlaceholderSegment) Render(hydrator Hydrator) []byte {
	return hydrator(s.id)
}
