package domain

// Category represents the category of a classified ad.
type Category string

const (
	CategoryImmo          Category = "immo"
	CategoryAuto          Category = "auto"
	CategoryConsumerGoods Category = "consumer_goods"
	CategoryHolidays      Category = "holidays"
)

// NewCategory validates and builds a Category from a raw string.
func NewCategory(s string) (Category, error) {
	switch Category(s) {
	case CategoryImmo, CategoryAuto, CategoryConsumerGoods, CategoryHolidays:
		return Category(s), nil
	default:
		return "", ErrInvalidCategory
	}
}
