package main

type Role struct {
	Base string
	Name string
}

func (r Role) URL() string {
	return r.Base + r.Name
}

func (r Role) Simple() string {
	return r.Name
}
