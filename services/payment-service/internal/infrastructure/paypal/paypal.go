package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Sed-Miyuki/OmniRoute/services/payment-service/internal/domain"
	"github.com/Sed-Miyuki/OmniRoute/services/payment-service/pkg/types"
)

type paypalClient struct {
	config     *types.PaymentConfig
	httpClient *http.Client
	baseURL    string
}

func NewPayPalClient(config *types.PaymentConfig) domain.PaymentProcessor {
	return &paypalClient{
		config:     config,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api-m.sandbox.paypal.com",
	}
}

func (p *paypalClient) getAccessToken(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v1/oauth2/token", p.baseURL),
		strings.NewReader("grant_type=client_credentials"),
	)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(p.config.PayPalClientID, p.config.PayPalSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("paypal auth failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func (p *paypalClient) CreatePaymentSession(
	ctx context.Context,
	amount int64,
	currency string,
	metadata map[string]string,
) (string, error) {
	token, err := p.getAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get paypal access token: %w", err)
	}

	payload := map[string]any{
		"intent": "CAPTURE",
		"purchase_units": []map[string]any{
			{
				"description": "Ride Payment",
				"amount": map[string]string{
					"currency_code": strings.ToUpper(currency),
					"value":         fmt.Sprintf("%.2f", float64(amount)/100.0),
				},
				"custom_id": metadata["trip_id"],
			},
		},
		"payment_source": map[string]any{
			"paypal": map[string]any{
				"experience_context": map[string]string{
					"return_url": p.config.ReturnURL,
					"cancel_url": p.config.CancelURL,
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/v2/checkout/orders", p.baseURL),
		bytes.NewBuffer(bodyBytes),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create paypal order (%d): %s", resp.StatusCode, string(body))
	}

	var orderResp struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orderResp); err != nil {
		return "", err
	}

	return orderResp.ID, nil
}
