package mcp

import (
	_ "embed"
	"encoding/json"
)

//go:embed catalog.json
var catalogJSON []byte

// CatalogItem is one curated starter. Items with a RegistryName can be
// refreshed from the official registry; the embedded spec is the offline copy.
type CatalogItem struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Category     string     `json:"category"`
	Description  string     `json:"description"`
	RegistryName string     `json:"registryName,omitempty"`
	Homepage     string     `json:"homepage,omitempty"`
	Spec         Spec       `json:"spec"`
	Inputs       []InputDef `json:"inputs,omitempty"`
	Notes        []string   `json:"notes,omitempty"`
}

// Catalog returns the embedded curated list.
func Catalog() []CatalogItem {
	var items []CatalogItem
	_ = json.Unmarshal(catalogJSON, &items)
	return items
}

// Draft turns a catalog item into a draft.
func (c CatalogItem) Draft() Draft {
	d := Draft{
		Name: c.ID, Title: c.Title, Description: c.Description, Spec: c.Spec.Normalize(),
		Inputs: c.Inputs, Origin: "catalog:" + c.ID, Notes: c.Notes, Homepage: c.Homepage,
	}
	if len(d.Inputs) == 0 {
		d.Inputs = detectInputs(d.Spec)
	}
	return d
}
