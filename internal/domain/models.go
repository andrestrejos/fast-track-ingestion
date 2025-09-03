package domain


type PaymentEvent struct {
UserID int `json:"user_id"`
PaymentID int `json:"payment_id"`
DepositAmount int `json:"deposit_amount"`
}