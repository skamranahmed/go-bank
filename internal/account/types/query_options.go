package types

type AccountQueryOptions struct {
	AccountID *int64
	Columns   []string

	// When true, the query will lock the selected row for update
	ForUpdate bool
}

type AccountUpdateOptions struct {
	NewBalance *int64
}
