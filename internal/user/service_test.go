package user

import (
	"context"
	"errors"
	"testing"

	"fitslot/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, name, email, passwordHash, role string) (*User, error) {
	args := m.Called(ctx, name, email, passwordHash, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) FindByID(ctx context.Context, id int) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func TestService_Register(t *testing.T) {
	tests := []struct {
		name          string
		req           RegisterRequest
		setupMock     func(*MockRepository)
		expectError   bool
		expectedError error
	}{
		{
			name: "successful registration",
			req: RegisterRequest{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("EmailExists", mock.Anything, "test@example.com").Return(false, nil)
				m.On("Create", mock.Anything, "Test User", "test@example.com", mock.Anything, "member").Return(&User{
					ID:    1,
					Name:  "Test User",
					Email: "test@example.com",
					Role:  "member",
				}, nil)
			},
			expectError: false,
		},
		{
			name: "email already exists",
			req: RegisterRequest{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("EmailExists", mock.Anything, "existing@example.com").Return(true, nil)
			},
			expectError:   true,
			expectedError: ErrEmailExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			service := NewService(mockRepo, "test-secret")
			user, accessToken, refreshToken, err := service.Register(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				assert.Nil(t, user)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_Login(t *testing.T) {
	tests := []struct {
		name          string
		req           LoginRequest
		setupMock     func(*MockRepository)
		expectError   bool
		expectedError error
	}{
		{
			name: "successful login",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				// Generate a proper password hash for "password123"
				passwordHash, _ := auth.HashPassword("password123")
				m.On("FindByEmail", mock.Anything, "test@example.com").Return(&User{
					ID:           1,
					Email:        "test@example.com",
					PasswordHash: passwordHash,
					Role:         "member",
				}, nil)
			},
			expectError: false,
		},
		{
			name: "user not found",
			req: LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockRepository) {
				m.On("FindByEmail", mock.Anything, "notfound@example.com").Return(nil, errors.New("not found"))
			},
			expectError:   true,
			expectedError: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tt.setupMock(mockRepo)

			service := NewService(mockRepo, "test-secret")
			user, accessToken, refreshToken, err := service.Login(context.Background(), tt.req)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedError != nil {
					assert.Equal(t, tt.expectedError, err)
				}
				assert.Nil(t, user)
				assert.Empty(t, accessToken)
				assert.Empty(t, refreshToken)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, accessToken)
				assert.NotEmpty(t, refreshToken)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	mockRepo := new(MockRepository)
	mockRepo.On("FindByID", mock.Anything, 1).Return(&User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Role:  "member",
	}, nil)

	service := NewService(mockRepo, "test-secret")
	user, err := service.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	mockRepo.AssertExpectations(t)
}

