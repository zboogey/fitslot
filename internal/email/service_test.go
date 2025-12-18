package email

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"fitslot/internal/logger"
)

func TestMain(m *testing.M) {
	logger.Init()

	code := m.Run()
	os.Exit(code)
}

// Вспомогательная функция для создания тестового сервиса с мок Redis
func newTestService(rdb *redis.Client) *Service {
	return &Service{
		redis:    rdb,
		from:     "noreply@fitslot.com",
		fromName: "FitSlot Team",
		smtpHost: "smtp.test.com",
		smtpPort: "587",
		smtpUser: "test@example.com",
		smtpPass: "password",
	}
}

func TestSend(t *testing.T) {
	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	// Используем Regexp для игнорирования содержимого
	mock.Regexp().ExpectLPush("emails", `.*`).SetVal(1)

	svc := newTestService(db)

	err := svc.Send(ctx, "user@example.com", "User", "Hello", "Test body")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSendBookingConfirmation(t *testing.T) {
	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	mock.Regexp().ExpectLPush("emails", `.*`).SetVal(1)

	svc := newTestService(db)

	when := time.Now().Add(24 * time.Hour)
	err := svc.SendBookingConfirmation(ctx, "user@example.com", "User", "Yoga Class", "Room A", when)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSendReminder(t *testing.T) {
	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	mock.Regexp().ExpectLPush("emails", `.*`).SetVal(1)

	svc := newTestService(db)

	when := time.Now().Add(24 * time.Hour)
	err := svc.SendReminder(ctx, "user@example.com", "User", "Pilates", "Room B", when)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSendCancellation(t *testing.T) {
	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	mock.Regexp().ExpectLPush("emails", `.*`).SetVal(1)

	svc := newTestService(db)

	err := svc.SendCancellation(ctx, "user@example.com", "User", "Boxing", "Room C")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueueLength(t *testing.T) {
	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	// Мокируем LLEN команду
	mock.ExpectLLen("emails").SetVal(5)

	svc := newTestService(db)

	length := svc.QueueLength(ctx)
	assert.Equal(t, int64(5), length)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueueLengthZero(t *testing.T) {
	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	mock.ExpectLLen("emails").SetVal(0)

	svc := newTestService(db)

	length := svc.QueueLength(ctx)
	assert.Equal(t, int64(0), length)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSendError(t *testing.T) {
	db, mock := redismock.NewClientMock()
	ctx := context.Background()

	// Мокируем ошибку Redis
	mock.Regexp().ExpectLPush("emails", `.*`).SetErr(assert.AnError)

	svc := newTestService(db)

	err := svc.Send(ctx, "user@example.com", "User", "Hello", "Test body")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
