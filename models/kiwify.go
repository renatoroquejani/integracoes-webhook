package models

// KiwifyWebhook representa a estrutura do webhook da Kiwify
type KiwifyWebhook struct {
	OrderID         string `json:"order_id"`
	OrderRef        string `json:"order_ref"`
	OrderStatus     string `json:"order_status"`
	WebhookEventType string `json:"webhook_event_type"`
	WebhookEventID  string `json:"webhook_event_id"`
	CreatedAt       string `json:"created_at"`
	Commissions     struct {
		Currency           string `json:"currency"`
		CommissionedStores []struct {
			CustomName string `json:"custom_name"`
			Type      string `json:"type"`
			Value     string `json:"value"`
		} `json:"commissioned_stores"`
	} `json:"commissions"`
	Customer struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
	} `json:"customer"`
	Product struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Regular     bool   `json:"regular"`
		Quantity    int    `json:"quantity"`
		Price       string `json:"price"`
		OrderBumpID string `json:"order_bump_id"`
	} `json:"product"`
	Payment struct {
		Method       string `json:"method"`
		Installments int    `json:"installments"`
		ProcessorID  string `json:"processor_id"`
		Status      string `json:"status"`
		SafeStatus  string `json:"safe_status"`
		Currency    string `json:"currency"`
		Value       string `json:"value"`
	} `json:"payment"`
	Producer struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"producer"`
	Store struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"store"`
	TrackingData struct {
		Source      string `json:"source"`
		Medium      string `json:"medium"`
		Campaign    string `json:"campaign"`
		Content     string `json:"content"`
		Term        string `json:"term"`
		Identifier  string `json:"identifier"`
		SegmentName string `json:"segment_name"`
		UTMSource   string `json:"utm_source"`
		UTMMedium   string `json:"utm_medium"`
		UTMCampaign string `json:"utm_campaign"`
		UTMContent  string `json:"utm_content"`
		UTMTerm     string `json:"utm_term"`
	} `json:"tracking_data"`
	Subscription struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Plan   struct {
			Name        string `json:"name"`
			Frequency   string `json:"frequency"`
			Recurrences int    `json:"recurrences"`
		} `json:"plan"`
	} `json:"subscription"`
	SubscriptionID string `json:"subscription_id"`
	AccessURL      string `json:"access_url"`
}

// KiwifyAbandonedCart representa a estrutura do webhook de carrinho abandonado da Kiwify
type KiwifyAbandonedCart struct {
	CheckoutLink      string `json:"checkout_link"`
	Country          string `json:"country"`
	CNPJ             string `json:"cnpj"`
	Email            string `json:"email"`
	Name             string `json:"name"`
	Phone            string `json:"phone"`
	ProductID        string `json:"product_id"`
	ProductName      string `json:"product_name"`
	StoreID          string `json:"store_id"`
	SubscriptionPlan interface{} `json:"subscription_plan"`
}

// KiwifyResponse representa a estrutura da resposta do webhook
type KiwifyResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
