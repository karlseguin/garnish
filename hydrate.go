package garnish

import (
	"encoding/json"
	"errors"
	"gopkg.in/karlseguin/typed.v1"
	"io"
	"net/http"
	"time"
)

type HydrateLoader func(fragment ReferenceFragment) []byte

type Fragment interface {
	Render(runtime *Runtime) []byte
	Size() int
	serialize(serializer Serializer)
}

type LiteralFragment []byte

func (f LiteralFragment) Render(runtime *Runtime) []byte {
	return f
}

func (f LiteralFragment) serialize(serializer Serializer) {
	serializer.Write(f)
}

func (f LiteralFragment) Size() int {
	return len(f)
}

type ReferenceFragment struct {
	size int
	typed.Typed
}

func NewReferenceFragment(data []byte) (ReferenceFragment, error) {
	var ref map[string]interface{}
	if err := json.Unmarshal(data, &ref); err != nil {
		return ReferenceFragment{}, err
	}
	return ReferenceFragment{
		size:  len(data) + 100,
		Typed: typed.Typed(ref),
	}, nil
}

func (f ReferenceFragment) Render(runtime *Runtime) []byte {
	return runtime.HydrateLoader(f)
}

func (f ReferenceFragment) Size() int {
	return f.size
}

func (f ReferenceFragment) serialize(serializer Serializer) {
	b, _ := f.Typed.ToBytes("")
	serializer.Write(b)
}

type HydrateResponse struct {
	status    int
	size      int
	expires   time.Time
	header    http.Header
	fragments []Fragment
}

func NewHydraterResponse(status int, header http.Header, fragments []Fragment) *HydrateResponse {
	return &HydrateResponse{
		status:    status,
		header:    header,
		fragments: fragments,
	}
}

func (r *HydrateResponse) Write(runtime *Runtime, w io.Writer) {
	for _, fragment := range r.fragments {
		w.Write(fragment.Render(runtime))
	}
}

func (r *HydrateResponse) Status() int {
	return r.status
}

func (r *HydrateResponse) AddHeader(key, value string) Response {
	r.header.Set(key, value)
	return r
}

func (r *HydrateResponse) Header() http.Header {
	return r.header
}

func (r *HydrateResponse) Close() {}

func (r *HydrateResponse) ContentLength() int {
	return -1
}

func (r *HydrateResponse) Size() int {
	return r.size
}

func (r *HydrateResponse) Cached() bool {
	return r.expires != zero
}

func (r *HydrateResponse) Expires() time.Time {
	return r.expires
}

func (r *HydrateResponse) Expire(at time.Time) {
	r.expires = at
}

func (r *HydrateResponse) ToCacheable(expires time.Time) CachedResponse {
	r.expires = expires
	r.size = 300 + 200*len(r.header)
	for _, f := range r.fragments {
		r.size += f.Size()
	}
	return r
}

func (r *HydrateResponse) Serialize(serializer Serializer) error {
	serializer.WriteInt(r.status)
	serializeHeader(serializer, r.header)
	serializer.WriteInt(len(r.fragments))
	for _, fragment := range r.fragments {
		switch fragment.(type) {
		case LiteralFragment:
			serializer.WriteByte(1)
		case ReferenceFragment:
			serializer.WriteByte(2)
		default:
			return errors.New("unknown fragment type")
		}
		fragment.serialize(serializer)
	}
	return nil
}

func (r *HydrateResponse) Deserialize(deserializer Deserializer) error {
	r.status = deserializer.ReadInt()
	r.header = deserializerHeader(deserializer)
	count := deserializer.ReadInt()
	r.fragments = make([]Fragment, count)
	for i := 0; i < count; i++ {
		switch deserializer.ReadByte() {
		case 1:
			r.fragments[i] = LiteralFragment(deserializer.CloneBytes())
		case 2:
			b := deserializer.ReadBytes()
			r.fragments[i], _ = NewReferenceFragment(b)
		default:
			return errors.New("unknown fragment type")
		}
	}
	return nil
}
