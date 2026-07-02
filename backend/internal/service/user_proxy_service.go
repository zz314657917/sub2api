package service

import (
	"context"
	"fmt"
	"strings"
)

type UserProxyService struct {
	proxyRepo ProxyRepository
}

func NewUserProxyService(proxyRepo ProxyRepository) *UserProxyService {
	return &UserProxyService{proxyRepo: proxyRepo}
}

func (s *UserProxyService) List(ctx context.Context, userID int64) ([]Proxy, error) {
	if s == nil || s.proxyRepo == nil {
		return nil, fmt.Errorf("proxy repository not configured")
	}
	return s.proxyRepo.ListUserOwned(ctx, userID)
}

func (s *UserProxyService) Create(ctx context.Context, userID int64, req CreateProxyRequest) (*Proxy, error) {
	if s == nil || s.proxyRepo == nil {
		return nil, fmt.Errorf("proxy repository not configured")
	}
	if userID <= 0 {
		return nil, ErrUserAccountNotOwned
	}
	proxy := &Proxy{
		Name:        strings.TrimSpace(req.Name),
		Protocol:    strings.ToLower(strings.TrimSpace(req.Protocol)),
		Host:        strings.TrimSpace(req.Host),
		Port:        req.Port,
		Username:    strings.TrimSpace(req.Username),
		Password:    strings.TrimSpace(req.Password),
		OwnerUserID: &userID,
		Status:      StatusActive,
	}
	if err := s.proxyRepo.Create(ctx, proxy); err != nil {
		return nil, fmt.Errorf("create user proxy: %w", err)
	}
	return proxy, nil
}

func (s *UserProxyService) Update(ctx context.Context, userID, proxyID int64, req UpdateProxyRequest) (*Proxy, error) {
	if s == nil || s.proxyRepo == nil {
		return nil, fmt.Errorf("proxy repository not configured")
	}
	proxy, err := s.proxyRepo.GetUserOwnedByID(ctx, userID, proxyID)
	if err != nil {
		return nil, fmt.Errorf("get user proxy: %w", err)
	}

	if req.Name != nil {
		proxy.Name = strings.TrimSpace(*req.Name)
	}
	if req.Protocol != nil {
		proxy.Protocol = strings.ToLower(strings.TrimSpace(*req.Protocol))
	}
	if req.Host != nil {
		proxy.Host = strings.TrimSpace(*req.Host)
	}
	if req.Port != nil {
		proxy.Port = *req.Port
	}
	if req.Username != nil {
		proxy.Username = strings.TrimSpace(*req.Username)
	}
	if req.Password != nil {
		proxy.Password = strings.TrimSpace(*req.Password)
	}
	if req.Status != nil {
		proxy.Status = strings.ToLower(strings.TrimSpace(*req.Status))
	}
	proxy.OwnerUserID = &userID

	if err := s.proxyRepo.Update(ctx, proxy); err != nil {
		return nil, fmt.Errorf("update user proxy: %w", err)
	}
	return proxy, nil
}

func (s *UserProxyService) Delete(ctx context.Context, userID, proxyID int64) error {
	if s == nil || s.proxyRepo == nil {
		return fmt.Errorf("proxy repository not configured")
	}
	if _, err := s.proxyRepo.GetUserOwnedByID(ctx, userID, proxyID); err != nil {
		return fmt.Errorf("get user proxy: %w", err)
	}
	count, err := s.proxyRepo.CountUserOwnedAccountsByProxyID(ctx, userID, proxyID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrProxyInUse
	}
	return s.proxyRepo.Delete(ctx, proxyID)
}
