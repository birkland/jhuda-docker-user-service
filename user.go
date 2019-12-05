package main

import (
	"encoding/json"
	"io"
)

type User struct {
	ID          string   `json:"@id"`
	Type        string   `json:"@type,omitempty"`
	Context     string   `json:"@context,omitempty"`
	Username    string   `json:"username,omitempty"`
	Firstname   string   `json:"firstName,omitempty"`
	Middlename  string   `json:"middleName,omitempty"`
	Lastname    string   `json:"lastName,omitempty"`
	Displayname string   `json:"displayName,omitempty"`
	Email       string   `json:"email,omitempty"`
	Affiliation []string `json:"affiliation,omitempty"`
	Locatorids  []string `json:"locatorIds,omitempty"`
	OrcidID     string   `json:"orcidId,omitempty"`
	Roles       []string `json:"roles,omitempty"`
}

func (u *User) Serialize(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(u)
}
