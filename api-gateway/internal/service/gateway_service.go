package service

import (
	"io"
	"net/http"
	"time"
)

type GatewayService struct {
	client *http.Client
}

func NewGatewayService() *GatewayService {
	return &GatewayService{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// ProxyRequest выполняет запрос к целевому URL, копируя тело и заголовки
func (s *GatewayService) ProxyRequest(method, url string, body io.Reader, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return s.client.Do(req)
}
