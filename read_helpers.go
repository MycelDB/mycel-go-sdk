package mycel

import "context"

func DoReadValue[T any](ctx context.Context, c *Client, operation string, fn func() (T, error)) (T, error) {
	return fn()
}

func DoRead(ctx context.Context, c *Client, operation string, fn func() error) error {
	return fn()
}
