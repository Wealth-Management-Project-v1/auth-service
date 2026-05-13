package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/wmp/auth-service/internal/config"
)

type PortfolioClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewPortfolioClient(cfg *config.Config) *PortfolioClient {
	return &PortfolioClient{
		baseURL: cfg.PortfolioServiceURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type createUserRequest struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	FullName string    `json:"fullName"`
}

func (c *PortfolioClient) CreateUser(id uuid.UUID, email, fullName string) error {
	body, err := json.Marshal(createUserRequest{
		ID:       id,
		Email:    email,
		FullName: fullName,
	})
	if err != nil {
		return fmt.Errorf("marshal create user request: %w", err)
	}

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/v1/users", bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("attempt %d: %w", attempt, err)
			slog.Warn("portfolio-service call failed, retrying", "attempt", attempt, "error", err)
			time.Sleep(time.Duration(attempt*attempt) * 100 * time.Millisecond)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			return nil
		}
		if resp.StatusCode == http.StatusConflict {
			// User already exists in portfolio-service, that's fine
			return nil
		}

		lastErr = fmt.Errorf("attempt %d: unexpected status %d", attempt, resp.StatusCode)
		slog.Warn("portfolio-service returned unexpected status", "attempt", attempt, "status", resp.StatusCode)
		time.Sleep(time.Duration(attempt*attempt) * 100 * time.Millisecond)
	}

	slog.Error("failed to create user in portfolio-service after retries", "error", lastErr)
	return lastErr
}
