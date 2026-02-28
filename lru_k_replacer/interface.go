package lrukreplacer

type ICache interface {
	Get(key int) (int, error)
	Set(key int, val int) (int, error)
}

type IEvict interface {
	evict() error
}
