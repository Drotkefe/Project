package service

import (
	"errors"
	"strings"
	"tripshare/internal/models"
	"tripshare/internal/repository"
)

type MemberService struct {
	repo    *repository.MemberRepository
	balance *BalanceService
}

func NewMemberService(repo *repository.MemberRepository, balance *BalanceService) *MemberService {
	return &MemberService{repo: repo, balance: balance}
}

func (s *MemberService) Create(req models.CreateMemberRequest) (*models.Member, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	member := &models.Member{Name: name}
	if err := s.repo.Create(member); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, errors.New("a member with this name already exists")
		}
		return nil, err
	}
	return member, nil
}

func (s *MemberService) List() ([]models.Member, error) {
	return s.repo.FindAll()
}

func (s *MemberService) Update(id uint, req models.UpdateMemberRequest) (*models.Member, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	member, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("member not found")
	}
	member.Name = name
	if err := s.repo.Update(member); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return nil, errors.New("a member with this name already exists")
		}
		return nil, err
	}
	return member, nil
}

func (s *MemberService) Delete(id uint) error {
	if _, err := s.repo.FindByID(id); err != nil {
		return errors.New("member not found")
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	return s.balance.RecalculateAll()
}
