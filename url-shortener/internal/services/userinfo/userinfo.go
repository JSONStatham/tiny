package userinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"urlshortener/internal/models"

	"github.com/mssola/useragent"
)

const ipAPIAddress = "http://ip-api.com/json"

type Service struct{}

type IPApiResponse struct {
	Status  string `json:"status"`
	Country string `json:"country"`
	Region  string `json:"regionName"`
	City    string `json:"city"`
}

type UserAgent struct {
	OS, Device, Browser string
}

func (s *Service) GetGeoInfo(ip string) (*IPApiResponse, error) {
	const op = "service.userinfo.GetGeoInfo"

	resp, err := http.Get(fmt.Sprintf("%s/%s", ipAPIAddress, ip))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer resp.Body.Close()

	var data *IPApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return data, nil
}

func (s *Service) ParseUserAgent(user_agent string) UserAgent {
	ua := useragent.New(user_agent)
	browser, _ := ua.Browser()
	os := ua.OS()

	device := "Desktop"
	if ua.Mobile() {
		device = "Mobile"
	} else if ua.Bot() {
		device = "Bot"
	}

	return UserAgent{
		OS:      os,
		Device:  device,
		Browser: browser,
	}
}

func (s *Service) ExtractRequestMeta(ctx context.Context, r *http.Request) (models.RequestMeta, error) {
	ip := getIP(r)
	userAgent := r.UserAgent()
	referrer := r.Referer()

	geoInfo, err := s.GetGeoInfo(ip)
	if err != nil {
		return models.RequestMeta{}, err
	}

	ua := s.ParseUserAgent(userAgent)

	return models.RequestMeta{
		IPAddress:  ip,
		UserAgent:  userAgent,
		Referrer:   referrer,
		Country:    geoInfo.Country,
		Region:     geoInfo.Region,
		City:       geoInfo.City,
		Browser:    ua.Browser,
		OS:         ua.OS,
		DeviceType: ua.Device,
	}, nil
}

func getIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
