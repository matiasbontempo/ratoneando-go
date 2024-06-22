package product

type Schema struct {
	ID        string  `json:"id"`
	Source    string  `json:"source,omitempty"`
	Name      string  `json:"name,omitempty"`
	Link      string  `json:"link,omitempty"`
	Image     string  `json:"image,omitempty"`
	Price     float64 `json:"price,omitempty"`
	Unit      string  `json:"unit,omitempty"`
	UnitPrice float64 `json:"unitPrice,omitempty"`
}

type ExtendedSchema struct {
	ID          string  `json:"id"`
	Source      string  `json:"source,omitempty"`
	Name        string  `json:"name,omitempty"`
	Link        string  `json:"link,omitempty"`
	Image       string  `json:"image,omitempty"`
	Unavailable bool    `json:"unavailable,omitempty"`
	Price       float64 `json:"price,omitempty"`
	ListPrice   float64 `json:"listPrice,omitempty"`
	Unit        string  `json:"unit,omitempty"`
	UnitPrice   float64 `json:"unitPrice,omitempty"`
	UnitFactor  float64 `json:"unitFactor,omitempty"`
}
