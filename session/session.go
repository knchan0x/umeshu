package session

// Session stores session values.
type Session interface {
	Get(key interface{}) interface{}  // gets session value
	Set(key, value interface{}) error // sets session value
	Delete(key interface{}) error     // deletes session value
}

// Default implementation of Session interface,
// reference type
type session map[interface{}]interface{}

var _ Session = (session)(nil) // interface check

// Get returns the value.
func (s session) Get(key interface{}) interface{} {
	if v, ok := s[key]; ok {
		return v
	}
	return nil
}

// Set sets key value pair.
func (s session) Set(key, value interface{}) error {
	s[key] = value
	return nil
}

// Delete deletes key value pair
func (s session) Delete(key interface{}) error {
	delete(s, key)
	return nil
}
