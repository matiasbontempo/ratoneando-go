package products

// PriceHistory is a small summary of how a product's price has moved. It is only present when
// there is enough data to say something.
type PriceHistory struct {
	Points     int     `json:"points"`
	WindowDays int     `json:"windowDays"`
	Typical    float64 `json:"typical"`
	Low        float64 `json:"low"`
	High       float64 `json:"high"`
	ChangePct  float64 `json:"changePct"`
	ChangeDays int     `json:"changeDays"`
}

type Schema struct {
	ID        string  `json:"id"`
	Source    string  `json:"source,omitempty"`
	Name      string  `json:"name,omitempty"`
	Link      string  `json:"link,omitempty"`
	Image     string  `json:"image,omitempty"`
	Price     float64 `json:"price,omitempty"`
	Unit      string  `json:"unit,omitempty"`
	UnitPrice float64 `json:"unitPrice,omitempty"`

	History *PriceHistory `json:"history,omitempty"`

	// Internal fields, used to record price history. They are not part of the public response.
	ListPrice  float64 `json:"-"`
	UnitFactor float64 `json:"-"`
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
