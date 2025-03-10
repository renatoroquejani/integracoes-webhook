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

// Configurações para OAuth do Meta Ads
const (
	metaAppID       = "647870654327010"                         // Substitua pelo App ID real do seu aplicativo Meta
	metaAppSecret   = "APP_SECRET"                              // Substitua pelo App Secret real do seu aplicativo Meta
	metaRedirectURI = "http://localhost:8081/meta-ads/callback" // URI de redirecionamento após autorização
	metaState       = "csrf_protection_state_value"             // Valor para proteção CSRF
)

// @title API de Webhooks e Integrações
// @version 1.0
// @description API para receber webhooks da Kiwify, Hotmart, Kirvano e integração com Meta Ads e Google Ads
// @host localhost:8081
// @BasePath /
// @tag.name Meta Ads
// @tag.description Endpoints para integração com Meta Ads
// @tag.name Google Ads
// @tag.description Endpoints para integração com Google Ads
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
	r.GET("/meta-ads/metricas", getMetaAdsMetricas)      // Suporte para GET
	r.POST("/api/meta-ads/metricas", getMetaAdsMetricas) // Rota adicional para compatibilidade com Swagger
	r.GET("/api/meta-ads/metricas", getMetaAdsMetricas)  // Suporte para GET na rota do Swagger
	r.GET("/meta-ads/campanha/:campaign_id", getMetaAdsCampaignInsights)
	r.GET("/meta-ads/conta/:account_id", getMetaAdsAccountInsights)

	// Novas rotas para autenticação OAuth do Meta Ads
	r.GET("/meta-ads/auth", handleMetaAdsAuth)
	r.GET("/meta-ads/callback", handleMetaAdsCallback)

	// Rotas para integração com Google Ads
	r.POST("/google-ads/metricas", getGoogleAdsMetricas)
	r.GET("/google-ads/metricas", getGoogleAdsMetricas)      // Suporte para GET
	r.POST("/api/google-ads/metricas", getGoogleAdsMetricas) // Rota adicional para compatibilidade com Swagger
	r.GET("/api/google-ads/metricas", getGoogleAdsMetricas)  // Suporte para GET na rota do Swagger
	r.GET("/google-ads/campanha/:campaign_id", getGoogleAdsCampaignInsights)
	r.GET("/google-ads/conta/:account_id", getGoogleAdsAccountInsights)
	r.GET("/google-ads/campanhas/:account_id", getGoogleAdsCampaigns)

	// Servir a página de demonstração do Meta Ads
	r.StaticFile("/meta-ads-demo.html", "./meta-ads-demo.html")

	// Servir a página de demonstração do Google Ads
	r.StaticFile("/google-ads-demo.html", "./google-ads-demo.html")

	// Servir arquivos estáticos em diretórios específicos
	r.Static("/static", "./static")

	// Inicia o servidor na porta 8081
	log.Println("Servidor iniciado em http://localhost:8081")
	log.Println("Swagger UI disponível em http://localhost:8081/swagger/index.html")
	log.Println("Meta Ads Demo disponível em http://localhost:8081/meta-ads-demo.html")
	log.Println("Google Ads Demo disponível em http://localhost:8081/google-ads-demo.html")
	log.Println("Meta Ads Auth disponível em http://localhost:8081/meta-ads/auth")

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

// @Summary Obter métricas do Google Ads
// @Description Obtém métricas como CTR, CPC, conversões e investimento total do Google Ads
// @Tags Google Ads
// @Accept json
// @Produce json
// @Param request body models.GoogleAdsRequest true "Credenciais de acesso do Google Ads"
// @Success 200 {object} models.GoogleAdsResponse
// @Failure 400 {object} models.GoogleAdsResponse
// @Failure 500 {object} models.GoogleAdsResponse
// @Router /google-ads/metricas [post]
func getGoogleAdsMetricas(c *gin.Context) {
	var request models.GoogleAdsRequest

	// Verificar se é uma requisição GET ou POST
	if c.Request.Method == "GET" {
		// Para GET, extrair parâmetros da query string
		request.ClientID = c.Query("client_id")
		request.ClientSecret = c.Query("client_secret")
		request.RefreshToken = c.Query("refresh_token")
		request.ManagerID = c.Query("manager_id")
	} else {
		// Para POST, extrair parâmetros do corpo da requisição
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			respondWithError(c, http.StatusBadRequest, "Erro ao ler payload")
			return
		}

		if err := json.Unmarshal(body, &request); err != nil {
			respondWithError(c, http.StatusBadRequest, "JSON inválido")
			return
		}
	}

	// Validar parâmetros obrigatórios
	if request.ClientID == "" || request.ClientSecret == "" || request.RefreshToken == "" {
		respondWithError(c, http.StatusBadRequest, "Credenciais incompletas")
		return
	}

	// Criar serviço do Google Ads
	googleAdsService := services.NewGoogleAdsService()

	// Obter métricas
	data, err := googleAdsService.GetMetricas(request.ClientID, request.ClientSecret, request.RefreshToken, request.ManagerID)
	if err != nil {
		// Tentar usar dados simulados em caso de erro
		mockResponse, _ := googleAdsService.FallbackToMockData(err)
		c.JSON(http.StatusOK, mockResponse)
		return
	}

	// Retornar resposta de sucesso
	response := models.GoogleAdsResponse{
		Success: true,
		Message: "Métricas obtidas com sucesso",
		Data:    *data,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Obter insights de campanha do Google Ads
// @Description Obtém insights detalhados de uma campanha específica do Google Ads
// @Tags Google Ads
// @Accept json
// @Produce json
// @Param campaign_id path string true "ID da campanha"
// @Param client_id query string true "ID do cliente OAuth"
// @Param client_secret query string true "Secret do cliente OAuth"
// @Param refresh_token query string true "Token de atualização OAuth"
// @Success 200 {object} models.GoogleAdsResponse
// @Failure 400 {object} models.GoogleAdsResponse
// @Failure 500 {object} models.GoogleAdsResponse
// @Router /google-ads/campanha/{campaign_id} [get]
func getGoogleAdsCampaignInsights(c *gin.Context) {
	// Extrair parâmetros
	campaignID := c.Param("campaign_id")
	clientID := c.Query("client_id")
	clientSecret := c.Query("client_secret")
	refreshToken := c.Query("refresh_token")

	// Validar parâmetros obrigatórios
	if campaignID == "" {
		respondWithError(c, http.StatusBadRequest, "ID da campanha não fornecido")
		return
	}

	if clientID == "" || clientSecret == "" || refreshToken == "" {
		respondWithError(c, http.StatusBadRequest, "Credenciais incompletas")
		return
	}

	// Criar serviço do Google Ads
	googleAdsService := services.NewGoogleAdsService()

	// Obter insights da campanha
	data, err := googleAdsService.GetCampaignInsights(clientID, clientSecret, refreshToken, campaignID)
	if err != nil {
		// Extrair informações detalhadas do erro
		errorInfo := extractErrorInfo(err)

		// Retornar resposta de erro
		response := models.GoogleAdsResponse{
			Success: false,
			Message: "Erro ao obter insights da campanha",
			Error:   errorInfo,
		}

		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Retornar resposta de sucesso
	response := models.GoogleAdsResponse{
		Success: true,
		Message: "Insights da campanha obtidos com sucesso",
		Data:    *data,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Obter insights de conta do Google Ads
// @Description Obtém insights detalhados de uma conta de anúncios do Google Ads
// @Tags Google Ads
// @Accept json
// @Produce json
// @Param account_id path string true "ID da conta de anúncios"
// @Param client_id query string true "ID do cliente OAuth"
// @Param client_secret query string true "Secret do cliente OAuth"
// @Param refresh_token query string true "Token de atualização OAuth"
// @Success 200 {object} models.GoogleAdsResponse
// @Failure 400 {object} models.GoogleAdsResponse
// @Failure 500 {object} models.GoogleAdsResponse
// @Router /google-ads/conta/{account_id} [get]
func getGoogleAdsAccountInsights(c *gin.Context) {
	// Extrair parâmetros
	accountID := c.Param("account_id")
	clientID := c.Query("client_id")
	clientSecret := c.Query("client_secret")
	refreshToken := c.Query("refresh_token")

	// Validar parâmetros obrigatórios
	if accountID == "" {
		respondWithError(c, http.StatusBadRequest, "ID da conta não fornecido")
		return
	}

	if clientID == "" || clientSecret == "" || refreshToken == "" {
		respondWithError(c, http.StatusBadRequest, "Credenciais incompletas")
		return
	}

	// Criar serviço do Google Ads
	googleAdsService := services.NewGoogleAdsService()

	// Obter insights da conta
	data, err := googleAdsService.GetAccountInsights(clientID, clientSecret, refreshToken, accountID)
	if err != nil {
		// Extrair informações detalhadas do erro
		errorInfo := extractErrorInfo(err)

		// Retornar resposta de erro
		response := models.GoogleAdsResponse{
			Success: false,
			Message: "Erro ao obter insights da conta",
			Error:   errorInfo,
		}

		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Retornar resposta de sucesso
	response := models.GoogleAdsResponse{
		Success: true,
		Message: "Insights da conta obtidos com sucesso",
		Data:    *data,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Listar campanhas do Google Ads
// @Description Lista todas as campanhas disponíveis para uma conta de anúncios do Google Ads
// @Tags Google Ads
// @Accept json
// @Produce json
// @Param account_id path string true "ID da conta de anúncios"
// @Param client_id query string true "ID do cliente OAuth"
// @Param client_secret query string true "Secret do cliente OAuth"
// @Param refresh_token query string true "Token de atualização OAuth"
// @Success 200 {object} models.GoogleAdsCampaignListResponse
// @Failure 400 {object} models.GoogleAdsCampaignListResponse
// @Failure 500 {object} models.GoogleAdsCampaignListResponse
// @Router /google-ads/campanhas/{account_id} [get]
func getGoogleAdsCampaigns(c *gin.Context) {
	// Extrair parâmetros
	accountID := c.Param("account_id")
	clientID := c.Query("client_id")
	clientSecret := c.Query("client_secret")
	refreshToken := c.Query("refresh_token")

	// Validar parâmetros obrigatórios
	if accountID == "" {
		respondWithError(c, http.StatusBadRequest, "ID da conta não fornecido")
		return
	}

	if clientID == "" || clientSecret == "" || refreshToken == "" {
		respondWithError(c, http.StatusBadRequest, "Credenciais incompletas")
		return
	}

	// Criar serviço do Google Ads
	googleAdsService := services.NewGoogleAdsService()

	// Listar campanhas
	campaigns, err := googleAdsService.ListCampaigns(clientID, clientSecret, refreshToken, accountID)
	if err != nil {
		// Extrair informações detalhadas do erro
		errorInfo := extractErrorInfo(err)

		// Retornar resposta de erro
		response := models.GoogleAdsCampaignListResponse{
			Success: false,
			Message: "Erro ao listar campanhas",
			Error:   errorInfo,
		}

		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Retornar resposta de sucesso
	response := models.GoogleAdsCampaignListResponse{
		Success: true,
		Message: "Campanhas listadas com sucesso",
		Data:    campaigns,
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

// @Summary Iniciar autenticação OAuth do Meta Ads
// @Description Redireciona o usuário para a página de autorização do Meta
// @Tags Meta Ads
// @Produce html
// @Success 302 {string} string "Redirecionamento para a página de autorização do Meta"
// @Failure 400 {object} models.MetaAdsResponse
// @Router /meta-ads/auth [get]
func handleMetaAdsAuth(c *gin.Context) {
	// Criar instância do serviço Meta Ads com as configurações
	metaService := services.NewMetaAdsServiceWithConfig(
		metaAppID,
		metaAppSecret,
		metaRedirectURI,
		metaState,
	)

	// Obter URL de autorização
	authURL, err := metaService.GetAuthURL()
	if err != nil {
		c.JSON(http.StatusBadRequest, models.MetaAdsResponse{
			Success: false,
			Message: "Erro ao gerar URL de autorização",
			Error: &models.ErrorInfo{
				Message: err.Error(),
				Type:    "OAuth Error",
			},
		})
		return
	}

	// Redirecionar o usuário para a URL de autorização
	c.Redirect(http.StatusFound, authURL)
}

// @Summary Callback da autenticação OAuth do Meta Ads
// @Description Recebe o código de autorização e o troca por um token de acesso
// @Tags Meta Ads
// @Produce json
// @Param code query string true "Código de autorização"
// @Param state query string true "Estado para verificação CSRF"
// @Success 200 {object} models.OAuthTokenResponse
// @Failure 400 {object} models.MetaAdsResponse
// @Router /meta-ads/callback [get]
func handleMetaAdsCallback(c *gin.Context) {
	// Obter código e state da query
	code := c.Query("code")
	state := c.Query("state")

	// Verificar se o state corresponde ao valor esperado (proteção CSRF)
	if state != metaState {
		c.JSON(http.StatusBadRequest, models.MetaAdsResponse{
			Success: false,
			Message: "Erro de validação: state inválido",
			Error: &models.ErrorInfo{
				Message: "O valor 'state' não corresponde ao valor esperado",
				Type:    "Security Error",
			},
		})
		return
	}

	// Verificar se o código foi fornecido
	if code == "" {
		c.JSON(http.StatusBadRequest, models.MetaAdsResponse{
			Success: false,
			Message: "Código de autorização não fornecido",
			Error: &models.ErrorInfo{
				Message: "O parâmetro 'code' é obrigatório",
				Type:    "Validation Error",
			},
		})
		return
	}

	// Criar instância do serviço Meta Ads com as configurações
	metaService := services.NewMetaAdsServiceWithConfig(
		metaAppID,
		metaAppSecret,
		metaRedirectURI,
		metaState,
	)

	// Trocar o código por um token de acesso
	tokenResponse, err := metaService.ExchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.MetaAdsResponse{
			Success: false,
			Message: "Erro ao obter token de acesso",
			Error: &models.ErrorInfo{
				Message: err.Error(),
				Type:    "OAuth Error",
			},
		})
		return
	}

	// Retornar o token de acesso
	c.JSON(http.StatusOK, tokenResponse)
}
