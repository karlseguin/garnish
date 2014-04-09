package ccache

type Configuration struct {
	maxItems       int
	buckets        int
	itemsToPrune   int
	deleteBuffer   int
	promoteBuffer  int
	getsPerPromote int32
}

func Configure() *Configuration {
	return &Configuration{
		buckets:        64,
		itemsToPrune:   500,
		deleteBuffer:   1024,
		getsPerPromote: 10,
		promoteBuffer:  1024,
		maxItems:       5000,
	}
}

func (c *Configuration) MaxItems(max int) *Configuration {
	c.maxItems = max
	return c
}

func (c *Configuration) Buckets(count int) *Configuration {
	c.buckets = count
	return c
}

func (c *Configuration) ItemsToPrune(count int) *Configuration {
	c.itemsToPrune = count
	return c
}

func (c *Configuration) PromoteBuffer(size int) *Configuration {
	c.promoteBuffer = size
	return c
}

func (c *Configuration) DeleteBuffer(size int) *Configuration {
	c.deleteBuffer = size
	return c
}

func (c *Configuration) GetsPerPromote(count int) *Configuration {
	c.getsPerPromote = int32(count)
	return c
}
