package types

type UserQueryOptions struct {
	Username *string
	Email    *string
	ID       *string
	Columns  []string
}

type UserUpdateOptions struct {
	Username       *string
	HashedPassword *string
}
