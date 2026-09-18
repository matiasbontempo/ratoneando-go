package vtex

type CoreProps struct {
	BaseUrl string
	Source  string
	Query   string
	Raw     bool
}

// ProductData holds the unit information VTEX ships as a stringified JSON
// inside the ProductData attribute of the catalog API.
type ProductData struct {
	MeasurementUnit string  `json:"MeasurementUnit"`
	UnitMultiplier  float64 `json:"UnitMultiplier"`
}

type CommertialOffer struct {
	Price                float64 `json:"Price"`
	ListPrice            float64 `json:"ListPrice"`
	PriceWithoutDiscount float64 `json:"PriceWithoutDiscount"`
	AvailableQuantity    int     `json:"AvailableQuantity"`
	IsAvailable          bool    `json:"IsAvailable"`
}

type Seller struct {
	SellerDefault   bool            `json:"sellerDefault"`
	CommertialOffer CommertialOffer `json:"commertialOffer"`
}

type Image struct {
	ImageUrl string `json:"imageUrl"`
}

type Item struct {
	Name            string   `json:"name"`
	Ean             string   `json:"ean"`
	MeasurementUnit string   `json:"measurementUnit"`
	UnitMultiplier  float64  `json:"unitMultiplier"`
	Images          []Image  `json:"images"`
	Sellers         []Seller `json:"sellers"`
}

type ResponseProduct struct {
	ProductId   string   `json:"productId"`
	ProductName string   `json:"productName"`
	Brand       string   `json:"brand"`
	Link        string   `json:"link"`
	LinkText    string   `json:"linkText"`
	ProductData []string `json:"ProductData"`
	Items       []Item   `json:"items"`
}

// ResponseStructure is the payload returned by the public catalog API,
// a plain array of products.
type ResponseStructure []ResponseProduct

type RawProduct struct {
	ResponseProduct
	ProductData
}
