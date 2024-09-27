package state

import (
	"errors"
	"sync"

	"github.com/exogmi/gossip/internal/models"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type UserManager struct {
	users map[string]*models.User // Key: nickname
	mu    sync.RWMutex
}

func NewUserManager() *UserManager {
	return &UserManager{
		users: make(map[string]*models.User),
	}
}

func (um *UserManager) AddUser(user *models.User) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if _, exists := um.users[user.Nickname]; exists {
		return ErrUserAlreadyExists
	}
	um.users[user.Nickname] = user
	return nil
}

func (um *UserManager) RemoveUser(nickname string) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if _, exists := um.users[nickname]; !exists {
		return ErrUserNotFound
	}
	delete(um.users, nickname)
	return nil
}

func (um *UserManager) GetUser(nickname string) (*models.User, error) {
	um.mu.RLock()
	defer um.mu.RUnlock()

	user, exists := um.users[nickname]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (um *UserManager) UpdateUser(user *models.User) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if _, exists := um.users[user.Nickname]; !exists {
		return ErrUserNotFound
	}
	um.users[user.Nickname] = user
	return nil
}

func (um *UserManager) ListUsers() []*models.User {
	um.mu.RLock()
	defer um.mu.RUnlock()

	users := make([]*models.User, 0, len(um.users))
	for _, user := range um.users {
		users = append(users, user)
	}
	return users
}

func (um *UserManager) UserExists(nickname string) bool {
	um.mu.RLock()
	defer um.mu.RUnlock()

	_, exists := um.users[nickname]
	return exists
}

func (um *UserManager) ChangeNickname(oldNick, newNick string) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	user, exists := um.users[oldNick]
	if !exists {
		return ErrUserNotFound
	}

	if _, exists := um.users[newNick]; exists {
		return ErrUserAlreadyExists
	}

	delete(um.users, oldNick)
	user.Nickname = newNick
	um.users[newNick] = user
	return nil
}
