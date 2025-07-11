package container

type Container[T any] struct {
	Value []T
}

func (c *Container[T]) GetValue() []T {
	return c.Value
}
