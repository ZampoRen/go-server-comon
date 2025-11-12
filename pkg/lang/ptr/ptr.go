package ptr

// Of 返回指向值的指针
func Of[T any](v T) *T {
	return &v
}

func From[T any](p *T) T {
	if p != nil {
		return *p
	}
	var t T
	return t
}

func FromOrDefault[T any](p *T, def T) T {
	if p != nil {
		return *p
	}
	return def
}
