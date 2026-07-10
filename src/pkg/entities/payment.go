package entities

// CheckoutInitRequest represents the request body to initiate a checkout
type CheckoutInitRequest struct {
	OrderAmount        float64 `json:"order_amount" binding:"required,gt=0"`
	OrderInvoiceNumber string  `json:"order_invoice_number" binding:"required"`
	OrderDescription   string  `json:"order_description" binding:"required"`
	PaymentMethod      string  `json:"payment_method"` // BANK_TRANSFER, CARD, etc.
	CustomerID         string  `json:"customer_id"`
	SuccessURL         string  `json:"success_url"`
	ErrorURL           string  `json:"error_url"`
	CancelURL          string  `json:"cancel_url"`
}

// CheckoutInitResponse represents the response containing the checkout info
type CheckoutInitResponse struct {
	CheckoutURL string            `json:"checkout_url"`
	FormValues  map[string]string `json:"form_values"`
}

// SePayIPNRequest represents the IPN (Instant Payment Notification) payload
// from SePay Payment Gateway.
type SePayIPNRequest struct {
	Timestamp        int64         `json:"timestamp"`
	NotificationType string        `json:"notification_type"` // ORDER_PAID
	Order            SePayIPNOrder `json:"order"`
	Transaction      SePayIPNTx    `json:"transaction"`
}

type SePayIPNOrder struct {
	ID                 string `json:"id"`
	OrderID            string `json:"order_id"`
	OrderStatus        string `json:"order_status"` // CAPTURED
	OrderCurrency      string `json:"order_currency"`
	OrderAmount        string `json:"order_amount"`
	OrderInvoiceNumber string `json:"order_invoice_number"`
	OrderDescription   string `json:"order_description"`
	UserAgent          string `json:"user_agent"`
	IPAddress          string `json:"ip_address"`
}

type SePayIPNTx struct {
	ID                   string  `json:"id"`
	PaymentMethod        string  `json:"payment_method"`
	TransactionID        string  `json:"transaction_id"`
	TransactionType      string  `json:"transaction_type"`
	TransactionDate      string  `json:"transaction_date"`
	TransactionStatus    string  `json:"transaction_status"` // APPROVED
	TransactionAmount    string  `json:"transaction_amount"`
	TransactionCurrency  string  `json:"transaction_currency"`
	AuthenticationStatus string  `json:"authentication_status"`
	CardNumber           *string `json:"card_number"`
	CardHolderName       *string `json:"card_holder_name"`
	CardExpiry           *string `json:"card_expiry"`
	CardBrand            *string `json:"card_brand"`
}

// SePayWebhookRequest represents the OLD bank transfer webhook payload.
// Kept for backward compatibility with SePay bank account webhooks.
type SePayWebhookRequest struct {
	ID              int     `json:"id"`
	Gateway         string  `json:"gateway"`
	TransactionDate string  `json:"transactionDate"`
	AccountNumber   string  `json:"accountNumber"`
	Code            string  `json:"code"`
	Content         string  `json:"content"`
	TransferType    string  `json:"transferType"`
	TransferAmount  float64 `json:"transferAmount"`
	Accumulated     float64 `json:"accumulated"`
	SubAccount      string  `json:"subAccount"`
	ReferenceCode   string  `json:"referenceCode"`
	Description     string  `json:"description"`
}
