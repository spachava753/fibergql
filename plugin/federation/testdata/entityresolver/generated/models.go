// Code generated by github.com/spachava753/fibergql, DO NOT EDIT.

package generated

type Hello struct {
	Name      string `json:"name"`
	Secondary string `json:"secondary"`
}

func (Hello) IsEntity() {}

type HelloWithErrors struct {
	Name string `json:"name"`
}

func (HelloWithErrors) IsEntity() {}

type MultiHello struct {
	Name string `json:"name"`
}

func (MultiHello) IsEntity() {}

type MultiHelloByNamesInput struct {
	Name string `json:"Name"`
}

type MultiHelloWithError struct {
	Name string `json:"name"`
}

func (MultiHelloWithError) IsEntity() {}

type MultiHelloWithErrorByNamesInput struct {
	Name string `json:"Name"`
}

type PlanetRequires struct {
	Name     string `json:"name"`
	Size     int    `json:"size"`
	Diameter int    `json:"diameter"`
}

func (PlanetRequires) IsEntity() {}

type PlanetRequiresNested struct {
	Name  string `json:"name"`
	World *World `json:"world"`
	Size  int    `json:"size"`
}

func (PlanetRequiresNested) IsEntity() {}

type World struct {
	Foo   string `json:"foo"`
	Bar   int    `json:"bar"`
	Hello *Hello `json:"hello"`
}

func (World) IsEntity() {}

type WorldName struct {
	Name string `json:"name"`
}

func (WorldName) IsEntity() {}

type WorldWithMultipleKeys struct {
	Foo   string `json:"foo"`
	Bar   int    `json:"bar"`
	Hello *Hello `json:"hello"`
}

func (WorldWithMultipleKeys) IsEntity() {}
