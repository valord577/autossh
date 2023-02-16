package component

type component interface {
	init() error
	free() error
}

func Use(c component) error {
	return c.init()
}

func Free(c component) error {
	return c.free()
}
