package authinfo

import "database/sql"

type AuthInfo struct {
	Account *AccountInfo `json:"account,omitempty"`
}

type AccountInfo struct {
	ID        int64        `json:"id"`
	Username  string       `json:"username"`
	Email     string       `json:"email"`
	CreatedAt sql.NullTime `json:"created_at"`
}
