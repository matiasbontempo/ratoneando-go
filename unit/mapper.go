package unit

const (
	kilo   = "KG"
	liters = "LT"
	units  = "UN"
	meters = "MT"
)

var unitMapper = map[string]string{
	"K":   kilo,
	"GR":  kilo,
	"G":   kilo,
	"L":   liters,
	"ML":  liters,
	"CC":  liters,
	"C":   liters,
	"M":   meters,
	"MI":  meters,
	"UD":  units,
	"UN":  units,
	"UNI": units,
	"U":   units,
}
