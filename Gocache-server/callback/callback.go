package callback

type CallBack interface {
	Get(key string) ([]byte, error)
}

type CallBackFunc func(key string) ([]byte, error)

func (f CallBackFunc) Get(key string) ([]byte, error) {
	return f(key)
}
