package store

type Store map[string]map[string]interface{}

func NewStore() Store {
	return make(Store)
}
