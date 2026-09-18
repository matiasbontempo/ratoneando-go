package unit

const (
	kilo   = "KG"
	liters = "LT"
	units  = "UN"
	meters = "MT"
)

var unitMapper = map[string]string{
	"K":   kilo,
	"KG":  kilo,
	"GR":  kilo,
	"G":   kilo,
	"L":   liters,
	"LT":  liters,
	"ML":  liters,
	"CC":  liters,
	"C":   liters,
	"M":   meters,
	"MT":  meters,
	"MI":  meters,
	"UD":  units,
	"UN":  units,
	"UNI": units,
	"U":   units,
}
