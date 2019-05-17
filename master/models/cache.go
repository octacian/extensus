package models

var cache = make(map[interface{}]Cacheable)

// Cacheable is any type containing a method to refresh the contents of the
// instance given an identifier.
type Cacheable interface {
	Refresh(interface{}) error
}

// Cache takes an empty instance of a cacheable items and an identifier for
// the wanted instance, returning a cached instance or attempting to fetch
// the item if it is not cached. If an error occurs it is returned.
func Cache(item Cacheable, identifier interface{}) (Cacheable, error) {
	if cached, ok := cache[identifier]; ok {
		return cached, nil
	}

	if err := item.Refresh(identifier); err != nil {
		return nil, err
	}

	cache[identifier] = item
	return item, nil
}
