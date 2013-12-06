package caches

// Cache storage implementations must follow this interface
type Cache interface {
	// Get a value from the cache
	Get(key, vary string) *CachedResponse

	// Store a value in the cache
	Set(key, vary string, response *CachedResponse)

	// Remove a value from the cache, including all its variances
	Delete(key string) bool

	// Remove a specific key+vary from the cache
	DeleteVary(key, vary string) bool
}
