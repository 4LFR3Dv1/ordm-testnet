package validator

type Transaction struct {
    From   string
    To     string
    Amount int
}

type Block struct {
    ID          string
    ParentID    string
    Transactions []Transaction
}
