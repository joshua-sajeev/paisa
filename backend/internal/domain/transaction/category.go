package transaction

type TransactionCategory string

const (
	TransactionCategoryFood          TransactionCategory = "food"
	TransactionCategoryTransport     TransactionCategory = "transport"
	TransactionCategoryEntertainment TransactionCategory = "entertainment"
	TransactionCategoryGroceries     TransactionCategory = "groceries"
	TransactionCategoryHealth        TransactionCategory = "health"
	TransactionCategoryTransfer      TransactionCategory = "transfer"
	TransactionCategoryDonation      TransactionCategory = "donation"
	TransactionCategoryInvestment    TransactionCategory = "investment"
	TransactionCategoryHousing       TransactionCategory = "housing"
	TransactionCategoryOther         TransactionCategory = "other"
)

func (tc TransactionCategory) IsValid() bool {
	switch tc {
	case TransactionCategoryFood,
		TransactionCategoryTransport,
		TransactionCategoryEntertainment,
		TransactionCategoryGroceries,
		TransactionCategoryHealth,
		TransactionCategoryTransfer,
		TransactionCategoryDonation,
		TransactionCategoryInvestment,
		TransactionCategoryHousing,
		TransactionCategoryOther:
		return true
	default:
		return false
	}
}
