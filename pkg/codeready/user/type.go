package codeready

type User struct {
	Username    string       `json:"username"`
	Enabled     bool         `json:"enabled"`
	Email       string       `json:"email"`
	Credentials []Credential `json:"credentials"`
	ClientRoles ClientRoles  `json:"clientRoles"`
}

type Credential struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ClientRoles struct {
	RealmManagement []string `json:"realm-management"`
}
