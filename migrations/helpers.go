package migrations

func pointer[T any](v T) *T {
	return &v
}
