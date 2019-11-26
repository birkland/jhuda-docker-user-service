package main

import (
	"encoding/json"
	"io"
)

type User struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Firstname   string   `json:"firstName"`
	Middlename  string   `json:"middleName"`
	Lastname    string   `json:"lastName"`
	Displayname string   `json:"displayName"`
	Email       string   `json:"email"`
	Affiliation []string `json:"affiliation"`
	Locatorids  []string `json:"locatorIds"`
	OrcidID     string   `json:"orcidId"`
	Roles       []string `json:"roles"`
}

func (u *User) Serialize(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(u)
}
