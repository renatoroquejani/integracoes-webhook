# POC Integrações Webhook

Este é um projeto de prova de conceito para integração de webhooks de diversos serviços como Kiwify, Hotmart e Kirvano.

## Requisitos

- Go 1.21 ou superior
- Git

## Como executar

1. Clone o repositório
   ```bash
   git clone https://github.com/renatoroquejani/integracoes-webhook.git
   cd integracoes-webhook
   ```

2. Execute o comando para baixar as dependências:
   ```bash
   go mod download
   ```

3. Execute a aplicação:
   ```bash
   go run main.go
   ```

O servidor estará rodando em http://localhost:8080

## Documentação da API

A documentação da API está disponível via Swagger UI em:
http://localhost:8080/swagger/index.html

## Endpoints disponíveis

### POST /webhook/kiwify

Recebe webhooks da Kiwify. Suporta tanto notificações de compra quanto de carrinho abandonado.

Parâmetros:
- Query: `signature` (obrigatório) - Assinatura para validação
- Body: Payload do webhook no formato JSON

Exemplo de uso com curl:
```bash
# Para uma compra aprovada
curl -X POST "http://localhost:8080/webhook/kiwify?signature=6ae35bad036fda55be4b27d5e9a93dd6f0992b62" \
  -H "Content-Type: application/json" \
  -d @payloads/kiwify/compra_aprovada.json

# Para um carrinho abandonado
curl -X POST "http://localhost:8080/webhook/kiwify?signature=2094d6f8b07871ad4a35f59b5f9051cf7c50f58e" \
  -H "Content-Type: application/json" \
  -d @payloads/kiwify/abandono_de_carrinho.json
```

### POST /webhook/hotmart

Recebe webhooks da Hotmart.

Exemplo de uso com curl:
```bash
curl -X POST http://localhost:8080/webhook/hotmart \
  -H "Content-Type: application/json" \
  -d @payloads/hotmart/compra_aprovada.json
```

### POST /webhook/kirvano

Recebe webhooks da Kirvano.

Exemplo de uso com curl:
```bash
curl -X POST http://localhost:8080/webhook/kirvano \
  -H "Content-Type: application/json" \
  -d @payloads/kirvano/compra_aprovada.json
```

## APIs de Integração com Plataformas de Anúncios

Esta API permite consultar dados de plataformas de anúncios como Meta Ads (Facebook/Instagram) e Google Ads usando tokens de acesso.

### Meta Ads (Facebook/Instagram)

As métricas disponíveis para Meta Ads incluem:

- CTR (Click-Through Rate)
- CAC (Custo de Aquisição por Cliente)
- Investimento Total
- Número de Vendas

#### Como obter um token de acesso do Meta Ads

Para utilizar esta API, você precisa de um token de acesso válido do Meta Ads. Siga os passos abaixo para obter um token:

1. Acesse o [Meta for Developers](https://developers.facebook.com/) e faça login com sua conta do Facebook
2. Crie um aplicativo no painel do desenvolvedor (Developer Dashboard)
3. Adicione o produto "Marketing API" ao seu aplicativo
4. Gere um token de acesso com as permissões: `ads_read`, `ads_management`, `business_management`
5. Para uso em produção, você precisará passar pelo processo de revisão do aplicativo

**Nota:** Para fins de teste, você pode usar um token de acesso temporário gerado no [Graph API Explorer](https://developers.facebook.com/tools/explorer/).

#### Endpoints do Meta Ads

##### POST /api/meta-ads/metricas

Consulta métricas específicas do Meta Ads usando um token fornecido no corpo da requisição.

**Exemplo de requisição:**

```json
{
  "token": "seu_token_de_acesso"
}
```

**Exemplo de resposta:**

```json
{
  "success": true,
  "message": "Métricas obtidas com sucesso",
  "data": {
    "id": "23847239847",
    "nome": "Campanha de Teste",
    "ctr": 3.2,
    "cac": 12.45,
    "investimento_total": 2500.75,
    "numero_vendas": 120
  }
}
```

##### GET /api/meta-ads/metricas

Consulta métricas específicas do Meta Ads usando um token fornecido via query parameter.

**Parâmetros:**

- `token` (obrigatório): Token de acesso ao Meta Ads
- `account_id` (opcional): ID da conta de anúncios
- `campaign_id` (opcional): ID da campanha

**Exemplo de requisição:**

```
GET /api/meta-ads/metricas?token=seu_token_de_acesso&campaign_id=123456789
```

**Exemplo de resposta:**

```json
{
  "success": true,
  "message": "Métricas obtidas com sucesso",
  "data": {
    "id": "123456789",
    "nome": "Campanha Específica",
    "ctr": 4.5,
    "cac": 9.75,
    "investimento_total": 1850.25,
    "numero_vendas": 95
  }
}
```

##### GET /api/meta-ads/consolidated

Consulta dados consolidados de todas as campanhas de todas as contas do Meta Ads.

**Parâmetros:**

- `token` (obrigatório): Token de acesso ao Meta Ads

**Exemplo de requisição:**

```
GET /api/meta-ads/consolidated?token=seu_token_de_acesso
```

**Exemplo de resposta:**

```json
{
  "success": true,
  "message": "Dados consolidados obtidos com sucesso",
  "data": [
    {
      "id": "act_123456789",
      "nome": "Conta Principal",
      "ctr": 2.8,
      "cac": 14.50,
      "investimento_total": 2230.25,
      "numero_vendas": 145
    },
    {
      "id": "23847239847",
      "nome": "Campanha de Conversão",
      "ctr": 3.2,
      "cac": 12.45,
      "investimento_total": 2500.75,
      "numero_vendas": 120
    },
    {
      "id": "23847239848",
      "nome": "Campanha de Reconhecimento",
      "ctr": 4.5,
      "cac": 9.75,
      "investimento_total": 1850.25,
      "numero_vendas": 95
    }
  ]
}
```

### Google Ads

As métricas disponíveis para Google Ads incluem:

- CTR (Click-Through Rate)
- CPC (Custo por Clique)
- Conversões
- Taxa de Conversão
- Custo por Conversão
- Investimento Total
- Impressões
- Cliques

#### Como obter credenciais do Google Ads

Para utilizar esta API, você precisa das seguintes credenciais do Google Ads:

1. Client ID
2. Client Secret
3. Refresh Token

Siga os passos abaixo para obter essas credenciais:

1. Acesse o [Google Cloud Console](https://console.cloud.google.com/)
2. Crie um projeto e configure o OAuth consent screen
3. Crie credenciais OAuth 2.0 para obter o Client ID e Client Secret
4. Use o fluxo de autorização OAuth para obter o Refresh Token

#### Endpoints do Google Ads

##### POST /api/google-ads/metricas

Consulta métricas específicas do Google Ads usando as credenciais fornecidas no corpo da requisição.

**Exemplo de requisição:**

```json
{
  "client_id": "seu_client_id",
  "client_secret": "seu_client_secret",
  "refresh_token": "seu_refresh_token",
  "manager_id": "id_da_conta_gerenciadora" // opcional
}
```

**Exemplo de resposta:**

```json
{
  "success": true,
  "message": "Métricas obtidas com sucesso",
  "data": {
    "id": "5808256042",
    "nome": "Conta 5808256042",
    "ctr": 2.08,
    "cpc": 24.83,
    "conversoes": 21,
    "taxa_conversao": 36.21,
    "custo_conversao": 68.58,
    "investimento_total": 1440.20,
    "impressions": 2789,
    "clicks": 58
  }
}
```

##### GET /api/google-ads/consolidated

Consulta dados consolidados de todas as campanhas de todas as contas do Google Ads.

**Parâmetros:**

- `client_id` (obrigatório): Client ID do OAuth
- `client_secret` (obrigatório): Client Secret do OAuth
- `refresh_token` (obrigatório): Refresh Token do OAuth

**Exemplo de requisição:**

```
GET /api/google-ads/consolidated?client_id=seu_client_id&client_secret=seu_client_secret&refresh_token=seu_refresh_token
```

**Exemplo de resposta:**

```json
{
  "success": true,
  "message": "Dados consolidados obtidos com sucesso",
  "data": [
    {
      "id": "5808256042",
      "nome": "Conta 5808256042",
      "ctr": 11.77,
      "cpc": 4.37,
      "conversoes": 50,
      "taxa_conversao": 2.86,
      "custo_conversao": 153.00,
      "investimento_total": 7650.15,
      "impressions": 14876,
      "clicks": 1751
    },
    {
      "id": "1",
      "nome": "Campanha Busca - Palavras-chave",
      "ctr": 2.08,
      "cpc": 24.83,
      "conversoes": 21,
      "taxa_conversao": 36.21,
      "custo_conversao": 68.58,
      "investimento_total": 1440.20,
      "impressions": 2789,
      "clicks": 58
    },
    {
      "id": "2",
      "nome": "Campanha Display - Interesses",
      "ctr": 2.08,
      "cpc": 24.83,
      "conversoes": 21,
      "taxa_conversao": 36.21,
      "custo_conversao": 68.58,
      "investimento_total": 1440.20,
      "impressions": 2789,
      "clicks": 58
    }
  ]
}
```

### Integração com APIs reais

Esta API se integra diretamente com as APIs oficiais do Meta Ads e Google Ads para obter dados reais das suas campanhas publicitárias. As integrações utilizam as bibliotecas oficiais:

- Meta Ads: [huandu/facebook](https://github.com/huandu/facebook) para Go
- Google Ads: Chamadas HTTP diretas para a API REST do Google Ads

As APIs realizam as seguintes operações:

1. Validam as credenciais fornecidas
2. Recuperam as contas de anúncios associadas às credenciais
3. Obtêm insights e métricas das campanhas ou contas de anúncios
4. Processam os dados para calcular métricas relevantes

## Exemplos de Dados Mockados (Apenas para Referência)

Os exemplos abaixo mostram o formato dos dados mockados que seriam retornados em caso de falha na API. Estes exemplos são apenas para referência e não são mais utilizados na aplicação.

### Meta Ads - Dados Mockados

```json
{
  "success": false,
  "message": "Usando dados simulados devido a falha na API. Detalhes do erro abaixo.",
  "data": {
    "id": "mock_campaign_123",
    "nome": "Campanha Simulada",
    "ctr": 2.5,
    "cac": 15.75,
    "investimento_total": 1250.50,
    "numero_vendas": 80
  },
  "error": {
    "message": "Token inválido ou não autorizado",
    "type": "AuthenticationError",
    "code": 190
  }
}
```

### Meta Ads - Dados Consolidados Mockados

```json
{
  "success": true,
  "message": "Dados consolidados obtidos com sucesso (usando dados simulados)",
  "data": [
    {
      "id": "mock_campaign_123",
      "nome": "Campanha Simulada 1",
      "ctr": 2.5,
      "cac": 15.75,
      "investimento_total": 1250.50,
      "numero_vendas": 80
    },
    {
      "id": "mock_campaign_456",
      "nome": "Campanha Simulada 2",
      "ctr": 3.2,
      "cac": 12.30,
      "investimento_total": 980.75,
      "numero_vendas": 65
    },
    {
      "id": "mock_account_789",
      "nome": "Conta Simulada",
      "ctr": 2.8,
      "cac": 14.50,
      "investimento_total": 2230.25,
      "numero_vendas": 145
    }
  ]
}
```

### Google Ads - Dados Mockados

```json
{
  "success": false,
  "message": "Usando dados simulados devido a falha na API. Detalhes do erro abaixo.",
  "data": {
    "id": "mock_campaign_123",
    "nome": "Campanha Simulada",
    "ctr": 2.5,
    "cpc": 1.75,
    "conversoes": 35,
    "taxa_conversao": 5.8,
    "custo_conversao": 21.50,
    "investimento_total": 752.50,
    "impressions": 7500,
    "clicks": 188
  },
  "error": {
    "message": "Credenciais inválidas ou não autorizadas",
    "type": "AuthenticationError",
    "code": 401
  }
}
```

### Google Ads - Dados Consolidados Mockados

```json
{
  "success": true,
  "message": "Dados consolidados obtidos com sucesso (usando dados simulados)",
  "data": [
    {
      "id": "123456789",
      "nome": "Conta 123456789",
      "ctr": 9.85,
      "cpc": 8.41,
      "conversoes": 38,
      "taxa_conversao": 3.18,
      "custo_conversao": 264.36,
      "investimento_total": 10045.75,
      "impressions": 12136,
      "clicks": 1195
    },
    {
      "id": "123456789_1",
      "nome": "Campanha 1",
      "ctr": 2.08,
      "cpc": 24.83,
      "conversoes": 21,
      "taxa_conversao": 36.21,
      "custo_conversao": 68.58,
      "investimento_total": 1440.20,
      "impressions": 2789,
      "clicks": 58
    },
    {
      "id": "high_performance_123",
      "nome": "Campanha de Alto Desempenho",
      "ctr": 8.5,
      "cpc": 1.25,
      "conversoes": 120,
      "taxa_conversao": 15.3,
      "custo_conversao": 8.75,
      "investimento_total": 1050.25,
      "impressions": 15000,
      "clicks": 1275
    }
  ]
}
```

## Estrutura do projeto

```
.
├── docs/               # Documentação Swagger
├── models/            # Modelos de dados
├── payloads/          # Exemplos de payloads
│   ├── hotmart/      # Payloads da Hotmart
│   ├── kiwify/       # Payloads da Kiwify
│   └── kirvano/      # Payloads da Kirvano
├── main.go           # Código principal
├── go.mod           # Dependências Go
└── README.md        # Este arquivo
```

## Contribuindo

1. Faça um fork do projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Faça commit das suas alterações (`git commit -am 'Adiciona nova feature'`)
4. Faça push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request
