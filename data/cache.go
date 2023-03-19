package data

type Cache interface { //Cache
	Len() int
	Get(string) (Value, bool)
	RemoveOldest()
	Add(string, Value)
	GetAll()
}
