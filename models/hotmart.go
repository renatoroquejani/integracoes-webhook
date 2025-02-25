package models

// HotmartWebhook representa a estrutura do webhook da Hotmart
type HotmartWebhook struct {
	Product struct {
		HasCoProduction bool   `json:"has_co_production"`
		Name           string `json:"name"`
		ID             int    `json:"id"`
		Ucode          string `json:"ucode,omitempty"` // Opcional em alguns eventos
	} `json:"product"`
	Commissions []struct {
		CurrencyValue string  `json:"currency_value"`
		Source        string  `json:"source"`
		Value         float64 `json:"value"`
	} `json:"commissions,omitempty"` // Opcional em abandono de carrinho
	Purchase *struct { // Opcional em abandono de carrinho
		Offer struct {
			Code string `json:"code"`
		} `json:"offer"`
		OrderDate          int64 `json:"order_date"`
		OriginalOfferPrice struct {
			CurrencyValue string  `json:"currency_value"`
			Value         float64 `json:"value"`
		} `json:"original_offer_price"`
		Price struct {
			CurrencyValue string  `json:"currency_value"`
			Value         float64 `json:"value"`
		} `json:"price"`
		CheckoutCountry struct {
			ISO  string `json:"iso"`
			Name string `json:"name"`
		} `json:"checkout_country"`
		SckPaymentLink string `json:"sckPaymentLink"`
		OrderBump     struct {
			ParentPurchaseTransaction string `json:"parent_purchase_transaction"`
			IsOrderBump              bool   `json:"is_order_bump"`
		} `json:"order_bump"`
		Payment struct {
			InstallmentsNumber int    `json:"installments_number"`
			Type              string `json:"type"`
		} `json:"payment"`
		ApprovedDate int64  `json:"approved_date"`
		FullPrice    struct {
			CurrencyValue string  `json:"currency_value"`
			Value         float64 `json:"value"`
		} `json:"full_price"`
		Transaction string `json:"transaction"`
		Status      string `json:"status"` // APPROVED, DISPUTE, EXPIRED, etc
	} `json:"purchase,omitempty"`
	Affiliates []struct {
		AffiliateCode string `json:"affiliate_code"`
		Name          string `json:"name"`
	} `json:"affiliates,omitempty"` // Opcional
	Producer *struct { // Opcional em abandono de carrinho
		Name string `json:"name"`
	} `json:"producer,omitempty"`
	Subscription *struct { // Opcional
		Subscriber struct {
			Code string `json:"code"`
		} `json:"subscriber"`
		Plan struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
		} `json:"plan"`
		Status string `json:"status"`
	} `json:"subscription,omitempty"`
	Buyer struct {
		Address *struct { // Opcional em abandono de carrinho
			Country    string `json:"country"`
			CountryISO string `json:"country_iso"`
		} `json:"address,omitempty"`
		Name          string `json:"name"`
		CheckoutPhone string `json:"checkout_phone,omitempty"` // Renomeado para phone em abandono de carrinho
		Phone         string `json:"phone,omitempty"`         // Usado em abandono de carrinho
		Email         string `json:"email"`
	} `json:"buyer"`
	BuyerIP   string `json:"buyer_ip,omitempty"` // Presente apenas em abandono de carrinho
	Affiliate bool   `json:"affiliate,omitempty"` // Presente apenas em abandono de carrinho
	Offer     *struct { // Presente apenas em abandono de carrinho
		Code string `json:"code"`
	} `json:"offer,omitempty"`
	Hottok       string `json:"hottok"`
	ID           string `json:"id"`
	CreationDate int64  `json:"creation_date"`
	Event        string `json:"event"`
	Version      string `json:"version"`
}

// HotmartResponse representa a estrutura da resposta do webhook
type HotmartResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
