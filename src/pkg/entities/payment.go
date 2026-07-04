package entities

// CheckoutInitRequest represents the request body to initiate a checkout
type CheckoutInitRequest struct {
	OrderAmount        float64 `json:"order_amount" binding:"required,gt=0"`
	OrderInvoiceNumber string  `json:"order_invoice_number" binding:"required"`
	OrderDescription   string  `json:"order_description" binding:"required"`
}

// CheckoutInitResponse represents the response containing the checkout info
type CheckoutInitResponse struct {
	CheckoutURL string            `json:"checkout_url"`
	FormValues  map[string]string `json:"form_values"`
}

// SePayWebhookRequest represents the payload that SePay sends to the webhook endpoint
type SePayWebhookRequest struct {
	ID             int     `json:"id"`
	Gateway        string  `json:"gateway"`
	TransactionDate string  `json:"transactionDate"`
	AccountNumber  string  `json:"accountNumber"`
	Code           string  `json:"code"`
	Content        string  `json:"content"`
	TransferType   string  `json:"transferType"`
	TransferAmount float64 `json:"transferAmount"`
	Accumulated    float64 `json:"accumulated"`
	SubAccount     string  `json:"subAccount"`
	ReferenceCode  string  `json:"referenceCode"`
	Description    string  `json:"description"`
}
