package models

// KirvanoWebhookBody representa a estrutura do body do webhook da Kirvano
type KirvanoWebhookBody struct {
	Event            string `json:"event"`
	EventDescription string `json:"event_description"`
	CheckoutID       string `json:"checkout_id"`
	SaleID           string `json:"sale_id"`         // Atualizado para sale_id
	PaymentMethod    string `json:"payment_method"`
	TotalPrice       string `json:"total_price"`
	Type             string `json:"type"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	Customer         struct {
		Name        string `json:"name"`
		Document    string `json:"document"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
	} `json:"customer"`
	Payment struct {
		Method       string `json:"method"`
		Brand        string `json:"brand"`
		Installments int    `json:"installments"`
		FinishedAt   string `json:"finished_at"`
	} `json:"payment"`
	Products []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		OfferID     string `json:"offer_id"`
		OfferName   string `json:"offer_name"`
		Description string `json:"description"`
		Price       string `json:"price"`
		Photo       string `json:"photo"`
		IsOrderBump bool   `json:"is_order_bump"`
	} `json:"products"`
	UTM struct {
		Src         string `json:"src"`
		UTMSource   string `json:"utm_source"`
		UTMMedium   string `json:"utm_medium"`
		UTMCampaign string `json:"utm_campaign"`
		UTMTerm     string `json:"utm_term"`
		UTMContent  string `json:"utm_content"`
	} `json:"utm"`
}

// KirvanoWebhook representa a estrutura completa do webhook da Kirvano
type KirvanoWebhook struct {
	Headers struct {
		Host           string `json:"host"`
		UserAgent      string `json:"user-agent"`
		ContentLength  string `json:"content-length"`
		Accept         string `json:"accept"`
		AcceptEncoding string `json:"accept-encoding"`
		Baggage        string `json:"baggage"`
		ContentType    string `json:"content-type"`
		NewRelic       string `json:"newrelic"`
		SentryTrace    string `json:"sentry-trace"`
		TraceParent    string `json:"traceparent"`
		TraceState     string `json:"tracestate"`
		XForwardedFor  string `json:"x-forwarded-for"`
		XForwardedHost string `json:"x-forwarded-host"`
		XForwardedPort string `json:"x-forwarded-port"`
		XForwardedProto string `json:"x-forwarded-proto"`
		XForwardedServer string `json:"x-forwarded-server"`
		XRealIP        string `json:"x-real-ip"`
	} `json:"headers"`
	Params     struct{} `json:"params"`
	Query      struct{} `json:"query"`
	Body       KirvanoWebhookBody `json:"body"`
	WebhookURL string `json:"webhookUrl"`
	ExecutionMode string `json:"executionMode"`
}

// KirvanoResponse representa a estrutura da resposta do webhook
type KirvanoResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
