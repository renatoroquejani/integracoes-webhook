package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	_ "poc-integracoes-onm/docs" // Importa a documentação do Swagger
	"poc-integracoes-onm/models"
	"poc-integracoes-onm/services"

	"github.com/gin-gonic/gin"
	fb "github.com/huandu/facebook/v2"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title API de Webhooks e Integrações
// @version 1.0
// @description API para receber webhooks da Kiwify, Hotmart, Kirvano e integração com Meta Ads
// @host localhost:8081
// @BasePath /
// @tag.name Meta Ads
// @tag.description Endpoints para integração com Meta Ads
func main() {
	// Configurar o modo de execução do Gin
	gin.SetMode(gin.ReleaseMode)

	// Criar uma instância do router Gin
	r := gin.New()

	// Adicionar middleware de recuperação e logger manualmente
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

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

	// Rotas para integração com Meta Ads
	r.POST("/meta-ads/metricas", getMetaAdsMetricas)
	r.GET("/meta-ads/metricas", getMetaAdsMetricas) // Suporte para GET
	r.POST("/api/meta-ads/metricas", getMetaAdsMetricas) // Rota adicional para compatibilidade com Swagger
	r.GET("/api/meta-ads/metricas", getMetaAdsMetricas)  // Suporte para GET na rota do Swagger
	r.GET("/meta-ads/campanha/:campaign_id", getMetaAdsCampaignInsights)
	r.GET("/meta-ads/conta/:account_id", getMetaAdsAccountInsights)

	// Servir a página de demonstração do Meta Ads
	r.StaticFile("/meta-ads-demo.html", "./meta-ads-demo.html")

	// Servir arquivos estáticos em diretórios específicos
	r.Static("/static", "./static")
	
	// Inicia o servidor na porta 8081
	log.Println("Servidor iniciado em http://localhost:8081")
	log.Println("Swagger UI disponível em http://localhost:8081/swagger/index.html")
	log.Println("Demo page disponível em http://localhost:8081/meta-ads-demo.html")

	// Iniciar o servidor
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
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

// Endpoint para obter métricas do Meta Ads
// @Summary Obter métricas do Meta Ads
// @Description Obtém métricas como CTR, CAC, investimento total e número de vendas do Meta Ads
// @Tags Meta Ads
// @Accept json
// @Produce json
// @Param request body models.MetaAdsRequest true "Token de acesso do Meta Ads"
// @Success 200 {object} models.MetaAdsResponse
// @Failure 400 {object} models.MetaAdsResponse
// @Failure 500 {object} models.MetaAdsResponse
// @Router /meta-ads/metricas [post]
func getMetaAdsMetricas(c *gin.Context) {
	// Variável para armazenar o token
	var token string

	// Verificar o método da requisição
	if c.Request.Method == "GET" {
		// Para GET, obter o token dos parâmetros de consulta
		token = c.Query("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, models.MetaAdsResponse{
				Success: false,
				Message: "Token não fornecido nos parâmetros de consulta",
				Error:   &models.ErrorInfo{Message: "Token é obrigatório"},
			})
			return
		}
	} else {
		// Para POST, obter o token do corpo da requisição
		var request models.MetaAdsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, models.MetaAdsResponse{
				Success: false,
				Message: "Erro ao processar a requisição",
				Error:   extractErrorInfo(err),
			})
			return
		}
		token = request.Token
	}

	// Criar o serviço Meta Ads
	metaAdsService := services.NewMetaAdsService()

	// Obter métricas do Meta Ads
	data, err := metaAdsService.GetMetricas(token)
	if err != nil {
		// Usar dados simulados em caso de erro, mas incluir detalhes do erro
		response, _ := metaAdsService.FallbackToMockData(err)
		c.JSON(http.StatusOK, response)
		return
	}

	// Retornar resposta de sucesso
	c.JSON(http.StatusOK, models.MetaAdsResponse{
		Success: true,
		Message: "Métricas obtidas com sucesso",
		Data:    *data,
	})
}

// Endpoint para obter insights de uma campanha específica do Meta Ads
// @Summary Obter insights de campanha do Meta Ads
// @Description Obtém insights detalhados de uma campanha específica do Meta Ads
// @Tags Meta Ads
// @Accept json
// @Produce json
// @Param campaign_id path string true "ID da campanha"
// @Param token query string true "Token de acesso do Meta Ads"
// @Success 200 {object} models.MetaAdsResponse
// @Failure 400 {object} models.MetaAdsResponse
// @Failure 500 {object} models.MetaAdsResponse
// @Router /meta-ads/campanha/{campaign_id} [get]
func getMetaAdsCampaignInsights(c *gin.Context) {
	campaignID := c.Param("campaign_id")
	token := c.Query("token")

	if campaignID == "" || token == "" {
		c.JSON(http.StatusBadRequest, models.MetaAdsResponse{
			Success: false,
			Message: "ID da campanha e token são obrigatórios",
			Error:   extractErrorInfo(nil),
		})
		return
	}

	// Criar o serviço Meta Ads
	metaAdsService := services.NewMetaAdsService()

	// Obter insights da campanha
	data, err := metaAdsService.GetCampaignInsights(token, campaignID)
	if err != nil {
		// Usar dados simulados em caso de erro, mas incluir detalhes do erro
		response, _ := metaAdsService.FallbackToMockData(err)
		c.JSON(http.StatusOK, response)
		return
	}

	// Retornar resposta de sucesso
	c.JSON(http.StatusOK, models.MetaAdsResponse{
		Success: true,
		Message: "Insights da campanha obtidos com sucesso",
		Data:    *data,
	})
}

// Endpoint para obter insights de uma conta de anúncios do Meta Ads
// @Summary Obter insights de conta do Meta Ads
// @Description Obtém insights detalhados de uma conta de anúncios do Meta Ads
// @Tags Meta Ads
// @Accept json
// @Produce json
// @Param account_id path string true "ID da conta de anúncios"
// @Param token query string true "Token de acesso do Meta Ads"
// @Success 200 {object} models.MetaAdsResponse
// @Failure 400 {object} models.MetaAdsResponse
// @Failure 500 {object} models.MetaAdsResponse
// @Router /meta-ads/conta/{account_id} [get]
func getMetaAdsAccountInsights(c *gin.Context) {
	accountID := c.Param("account_id")
	token := c.Query("token")

	if accountID == "" || token == "" {
		c.JSON(http.StatusBadRequest, models.MetaAdsResponse{
			Success: false,
			Message: "ID da conta e token são obrigatórios",
			Error:   extractErrorInfo(nil),
		})
		return
	}

	// Criar o serviço Meta Ads
	metaAdsService := services.NewMetaAdsService()

	// Obter insights da conta
	data, err := metaAdsService.GetAccountInsights(token, accountID)
	if err != nil {
		// Usar dados simulados em caso de erro, mas incluir detalhes do erro
		response, _ := metaAdsService.FallbackToMockData(err)
		c.JSON(http.StatusOK, response)
		return
	}

	// Retornar resposta de sucesso
	c.JSON(http.StatusOK, models.MetaAdsResponse{
		Success: true,
		Message: "Insights da conta obtidos com sucesso",
		Data:    *data,
	})
}

func respondWithError(c *gin.Context, code int, message string) {
	response := models.HotmartResponse{
		Status:  "error",
		Message: message,
	}
	c.JSON(code, response)
}

// extractErrorInfo extrai informações detalhadas de um erro
func extractErrorInfo(err error) *models.ErrorInfo {
	errorInfo := &models.ErrorInfo{
		Message: err.Error(),
	}
	
	// Verificar se é um erro do Facebook SDK
	if fbErr, ok := err.(*fb.Error); ok {
		errorInfo.Code = fbErr.Code
		errorInfo.Type = fbErr.Type
		errorInfo.Message = fbErr.Message
		errorInfo.Details = fbErr.TraceID
		
		// Tratar especificamente o erro de versão da API obsoleta
		if fbErr.Code == 2635 {
			errorInfo.Type = "DeprecatedAPIVersionError"
			errorInfo.Details = "A versão da API do Meta Ads utilizada está obsoleta. A aplicação será atualizada para usar a versão mais recente (v22.0)."
		}
	} else {
		// Tentar extrair mais informações do erro
		errorStr := err.Error()
		if strings.Contains(errorStr, "token inválido") {
			errorInfo.Type = "AuthenticationError"
			errorInfo.Code = 190
			errorInfo.Details = "O token fornecido é inválido ou expirou"
		} else if strings.Contains(errorStr, "permissão") {
			errorInfo.Type = "PermissionError"
			errorInfo.Code = 200
			errorInfo.Details = "O token não tem permissões suficientes"
		} else if strings.Contains(errorStr, "limite") {
			errorInfo.Type = "RateLimitError"
			errorInfo.Code = 4
			errorInfo.Details = "Limite de requisições excedido"
		} else if strings.Contains(errorStr, "deprecated version") || strings.Contains(errorStr, "2635") {
			errorInfo.Type = "DeprecatedAPIVersionError"
			errorInfo.Code = 2635
			errorInfo.Details = "A versão da API do Meta Ads utilizada está obsoleta. A aplicação será atualizada para usar a versão mais recente (v22.0)."
		}
	}
	
	return errorInfo
}
