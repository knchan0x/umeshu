package session

// Store is the persistence store for session.
type Store interface {
	// find session object by session id, return nil if no such
	// session id
	Read(sid string) (Session, error)

	// creates new session according to session id and token
	// and insert it into persistence store
	Insert(sid string, token string) (Session, error)

	// replaces old session id by new id
	UpdateSID(old string, new string)

	// deletes session according to session id
	Delete(sid string) error

	// force GC to remove all sessions excess maxLifeTime,
	// count in seconds
	GC(maxLifeTime int)
}

type constructor func(SessionSettings) Store

// storeConstructorMap stores all registered store type.
var storeConstructorMap = make(map[string]constructor)

// NewStore creates new store instance according to StoreType,
// returns nil if no such storeType.
func NewStore(storeType string, settings SessionSettings) Store {
	if f, ok := storeConstructorMap[storeType]; ok {
		return f(settings)
	}
	return nil
}

// Register registers storeType.
func Register(storeType string, constructor constructor) {
	storeConstructorMap[storeType] = constructor
}

// AvailableStoreTypes returns a slice of registered store types.
func AvailableStoreTypes() (list []string) {
	for storeType := range storeConstructorMap {
		list = append(list, storeType)
	}
	return list
}
