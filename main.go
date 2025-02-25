package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	_ "poc-integracoes-onm/docs" // Importa a documentação do Swagger
	"poc-integracoes-onm/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title API de Webhooks - Kiwify, Hotmart e Kirvano
// @version 1.0
// @description API para receber webhooks da Kiwify, Hotmart e Kirvano
// @host localhost:8080
// @BasePath /
func main() {
	r := gin.Default()

	// Rota do Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Rota de echo original
	r.POST("/webhook/echo", handleEcho)

	// Nova rota para webhook da Hotmart
	r.POST("/webhook/hotmart", handleHotmart)

	// Rota para webhook da Kiwify
	r.POST("/webhook/kiwify", handleKiwify)

	// Rota para webhook da Kirvano
	r.POST("/webhook/kirvano", handleKirvano)

	// Inicia o servidor na porta 8080
	log.Println("Servidor iniciado em http://localhost:8080")
	log.Println("Swagger UI disponível em http://localhost:8080/swagger/index.html")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

// @Summary Echo webhook
// @Description Retorna o mesmo payload recebido
// @Accept json
// @Produce json
// @Success 200 {object} interface{}
// @Router /webhook/echo [post]
func handleEcho(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao ler payload"})
		return
	}

	var payload interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	c.JSON(http.StatusOK, payload)
}

// @Summary Webhook Hotmart
// @Description Recebe notificações da Hotmart
// @Accept json
// @Produce json
// @Param webhook body models.HotmartWebhook true "Payload do webhook"
// @Success 200 {object} models.HotmartResponse
// @Failure 400 {object} models.HotmartResponse
// @Router /webhook/hotmart [post]
func handleHotmart(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Erro ao ler payload")
		return
	}

	var webhook models.HotmartWebhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		respondWithError(c, http.StatusBadRequest, "JSON inválido para webhook Hotmart")
		return
	}

	if webhook.Product.Ucode == "" {
		respondWithError(c, http.StatusBadRequest, "Ucode do produto não fornecido")
		return
	}

	if webhook.Purchase != nil && webhook.Purchase.Transaction == "" {
		respondWithError(c, http.StatusBadRequest, "Código da transação não fornecido")
		return
	}

	log.Printf("Novo evento Hotmart recebido: Transaction=%s, Product=%s, Status=%s\n",
		webhook.Purchase.Transaction,
		webhook.Product.Name,
		webhook.Purchase.Status)

	if len(webhook.Affiliates) > 0 {
		log.Printf("Afiliado presente na venda: %s (%s)\n",
			webhook.Affiliates[0].Name,
			webhook.Affiliates[0].AffiliateCode)
	}

	response := models.HotmartResponse{
		Status:  "success",
		Message: "Webhook processado com sucesso",
		Data:    webhook,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Webhook Kiwify
// @Description Recebe notificações da Kiwify
// @Accept json
// @Produce json
// @Param signature query string true "Assinatura para validação"
// @Param webhook body models.KiwifyWebhook true "Payload do webhook"
// @Success 200 {object} models.KiwifyResponse
// @Failure 400 {object} models.KiwifyResponse
// @Router /webhook/kiwify [post]
func handleKiwify(c *gin.Context) {
	// Verifica a assinatura
	signature := c.Query("signature")
	if signature == "" {
		respondWithError(c, http.StatusBadRequest, "Assinatura não fornecida")
		return
	}

	log.Printf("Assinatura recebida: %s\n", signature)

	// Lê o corpo da requisição
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Erro ao ler payload")
		return
	}

	log.Printf("Payload recebido: %s\n", string(body))

	// Tenta decodificar como webhook normal
	var webhook models.KiwifyWebhook
	err1 := json.Unmarshal(body, &webhook)

	// Se falhar, tenta decodificar como carrinho abandonado
	var abandonedCart models.KiwifyAbandonedCart
	err2 := json.Unmarshal(body, &abandonedCart)

	if err1 != nil && err2 != nil {
		log.Printf("Erro ao decodificar webhook: %v\n", err1)
		log.Printf("Erro ao decodificar carrinho abandonado: %v\n", err2)
		respondWithError(c, http.StatusBadRequest, "JSON inválido para webhook Kiwify")
		return
	}

	// Se for um webhook normal
	if err1 == nil && webhook.OrderID != "" {
		if webhook.OrderStatus == "" {
			respondWithError(c, http.StatusBadRequest, "Status do pedido não fornecido")
			return
		}

		log.Printf("Novo evento Kiwify recebido: OrderID=%s, Status=%s, Evento=%s\n",
			webhook.OrderID,
			webhook.OrderStatus,
			webhook.WebhookEventType)

		log.Printf("Cliente: Nome=%s, Email=%s\n",
			webhook.Customer.Name,
			webhook.Customer.Email)

		log.Printf("Produto: ID=%s, Nome=%s, Preço=%s\n",
			webhook.Product.ID,
			webhook.Product.Name,
			webhook.Product.Price)

		if webhook.TrackingData.UTMSource != "" {
			log.Printf("Origem: source=%s, medium=%s, campaign=%s\n",
				webhook.TrackingData.UTMSource,
				webhook.TrackingData.UTMMedium,
				webhook.TrackingData.UTMCampaign)
		}

		response := models.KiwifyResponse{
			Status:  "success",
			Message: "Webhook processado com sucesso",
			Data:    webhook,
		}

		c.JSON(http.StatusOK, response)
		return
	}

	// Se for um carrinho abandonado
	if err2 == nil && abandonedCart.CheckoutLink != "" {
		log.Printf("Novo evento de carrinho abandonado: Link=%s, Produto=%s\n",
			abandonedCart.CheckoutLink,
			abandonedCart.ProductName)

		log.Printf("Cliente: Nome=%s, Email=%s\n",
			abandonedCart.Name,
			abandonedCart.Email)

		response := models.KiwifyResponse{
			Status:  "success",
			Message: "Webhook de carrinho abandonado processado com sucesso",
			Data:    abandonedCart,
		}

		c.JSON(http.StatusOK, response)
		return
	}

	respondWithError(c, http.StatusBadRequest, "Payload inválido")
}

// @Summary Webhook Kirvano
// @Description Recebe notificações da Kirvano
// @Accept json
// @Produce json
// @Param webhook body models.KirvanoWebhookBody true "Payload do webhook"
// @Success 200 {object} models.KirvanoResponse
// @Failure 400 {object} models.KirvanoResponse
// @Router /webhook/kirvano [post]
func handleKirvano(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		respondWithError(c, http.StatusBadRequest, "Erro ao ler payload")
		return
	}

	log.Printf("Payload recebido: %s\n", string(body))

	var webhook models.KirvanoWebhookBody
	if err := json.Unmarshal(body, &webhook); err != nil {
		log.Printf("Erro ao decodificar payload: %v\n", err)
		respondWithError(c, http.StatusBadRequest, "JSON inválido para webhook Kirvano")
		return
	}

	if webhook.SaleID == "" {
		respondWithError(c, http.StatusBadRequest, "ID da venda não fornecido")
		return
	}

	if webhook.Status == "" {
		respondWithError(c, http.StatusBadRequest, "Status não fornecido")
		return
	}

	log.Printf("Novo evento Kirvano recebido: SaleID=%s, Status=%s, Evento=%s\n",
		webhook.SaleID,
		webhook.Status,
		webhook.Event)

	log.Printf("Cliente: Nome=%s, Email=%s\n",
		webhook.Customer.Name,
		webhook.Customer.Email)

	for _, product := range webhook.Products {
		log.Printf("Produto: ID=%s, Nome=%s, Preço=%s, OrderBump=%v\n",
			product.ID,
			product.Name,
			product.Price,
			product.IsOrderBump)
	}

	if webhook.UTM.Src != "" {
		log.Printf("Origem: src=%s, medium=%s, campaign=%s\n",
			webhook.UTM.Src,
			webhook.UTM.UTMMedium,
			webhook.UTM.UTMCampaign)
	}

	response := models.KirvanoResponse{
		Status:  "success",
		Message: "Webhook processado com sucesso",
		Data:    webhook,
	}

	c.JSON(http.StatusOK, response)
}

func respondWithError(c *gin.Context, code int, message string) {
	response := models.HotmartResponse{
		Status:  "error",
		Message: message,
	}
	c.JSON(code, response)
}
