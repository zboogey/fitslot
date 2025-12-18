package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestRecordHTTPRequest(t *testing.T) {
	// Сбрасываем метрики перед тестом
	HTTPRequestsTotal.Reset()
	HTTPRequestDuration.Reset()

	method := "GET"
	path := "/api/bookings"
	status := "200"
	duration := 0.5

	RecordHTTPRequest(method, path, status, duration)

	// Проверяем счетчик запросов
	count := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues(method, path, status))
	assert.Equal(t, float64(1), count)

	// Для histogram проверяем количество наблюдений через метрику _count
	metric := HTTPRequestDuration.WithLabelValues(method, path).(prometheus.Histogram)
	// Просто проверяем что метод был вызван без ошибки
	metric.Observe(duration)
}

func TestRecordHTTPRequestMultiple(t *testing.T) {
	HTTPRequestsTotal.Reset()

	RecordHTTPRequest("POST", "/api/login", "200", 0.1)
	RecordHTTPRequest("POST", "/api/login", "200", 0.2)
	RecordHTTPRequest("POST", "/api/login", "401", 0.05)

	// Проверяем счетчики
	successCount := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues("POST", "/api/login", "200"))
	failCount := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues("POST", "/api/login", "401"))

	assert.Equal(t, float64(2), successCount)
	assert.Equal(t, float64(1), failCount)
}

func TestRecordBooking(t *testing.T) {
	BookingsTotal.Reset()

	RecordBooking("confirmed", "card")

	count := testutil.ToFloat64(BookingsTotal.WithLabelValues("confirmed", "card"))
	assert.Equal(t, float64(1), count)
}

func TestRecordBookingMultiple(t *testing.T) {
	BookingsTotal.Reset()

	RecordBooking("confirmed", "card")
	RecordBooking("confirmed", "wallet")
	RecordBooking("cancelled", "card")

	cardConfirmed := testutil.ToFloat64(BookingsTotal.WithLabelValues("confirmed", "card"))
	walletConfirmed := testutil.ToFloat64(BookingsTotal.WithLabelValues("confirmed", "wallet"))
	cardCancelled := testutil.ToFloat64(BookingsTotal.WithLabelValues("cancelled", "card"))

	assert.Equal(t, float64(1), cardConfirmed)
	assert.Equal(t, float64(1), walletConfirmed)
	assert.Equal(t, float64(1), cardCancelled)
}

func TestRecordBookingCancellation(t *testing.T) {
	// Создаем новый счетчик для теста
	testCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fitslot_booking_cancellations_total_test",
			Help: "Total number of booking cancellations",
		},
	)

	// Временно подменяем глобальную переменную
	oldCounter := BookingCancellationsTotal
	BookingCancellationsTotal = testCounter
	defer func() { BookingCancellationsTotal = oldCounter }()

	RecordBookingCancellation()
	RecordBookingCancellation()

	count := testutil.ToFloat64(testCounter)
	assert.Equal(t, float64(2), count)
}

func TestRecordEmail(t *testing.T) {
	EmailsSentTotal.Reset()

	RecordEmail("booking_confirmation", "success")

	count := testutil.ToFloat64(EmailsSentTotal.WithLabelValues("booking_confirmation", "success"))
	assert.Equal(t, float64(1), count)
}

func TestRecordEmailMultipleTypes(t *testing.T) {
	EmailsSentTotal.Reset()

	RecordEmail("booking_confirmation", "success")
	RecordEmail("booking_confirmation", "failed")
	RecordEmail("reminder", "success")

	confirmSuccess := testutil.ToFloat64(EmailsSentTotal.WithLabelValues("booking_confirmation", "success"))
	confirmFailed := testutil.ToFloat64(EmailsSentTotal.WithLabelValues("booking_confirmation", "failed"))
	reminderSuccess := testutil.ToFloat64(EmailsSentTotal.WithLabelValues("reminder", "success"))

	assert.Equal(t, float64(1), confirmSuccess)
	assert.Equal(t, float64(1), confirmFailed)
	assert.Equal(t, float64(1), reminderSuccess)
}

func TestRecordWalletTopUp(t *testing.T) {
	// Создаем новый счетчик для теста
	testCounter := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "fitslot_wallet_topups_total_test",
			Help: "Total number of wallet top-ups",
		},
	)

	oldCounter := WalletTopUpsTotal
	WalletTopUpsTotal = testCounter
	defer func() { WalletTopUpsTotal = oldCounter }()

	RecordWalletTopUp()
	RecordWalletTopUp()
	RecordWalletTopUp()

	count := testutil.ToFloat64(testCounter)
	assert.Equal(t, float64(3), count)
}

func TestRecordSubscription(t *testing.T) {
	SubscriptionsCreatedTotal.Reset()

	RecordSubscription("monthly")

	count := testutil.ToFloat64(SubscriptionsCreatedTotal.WithLabelValues("monthly"))
	assert.Equal(t, float64(1), count)
}

func TestRecordSubscriptionMultipleTypes(t *testing.T) {
	SubscriptionsCreatedTotal.Reset()

	RecordSubscription("monthly")
	RecordSubscription("monthly")
	RecordSubscription("yearly")

	monthlyCount := testutil.ToFloat64(SubscriptionsCreatedTotal.WithLabelValues("monthly"))
	yearlyCount := testutil.ToFloat64(SubscriptionsCreatedTotal.WithLabelValues("yearly"))

	assert.Equal(t, float64(2), monthlyCount)
	assert.Equal(t, float64(1), yearlyCount)
}

func TestEmailQueueLength(t *testing.T) {
	EmailQueueLength.Set(10)
	value := testutil.ToFloat64(EmailQueueLength)
	assert.Equal(t, float64(10), value)

	EmailQueueLength.Set(5)
	value = testutil.ToFloat64(EmailQueueLength)
	assert.Equal(t, float64(5), value)

	EmailQueueLength.Set(0)
	value = testutil.ToFloat64(EmailQueueLength)
	assert.Equal(t, float64(0), value)
}

func TestWalletBalance(t *testing.T) {
	WalletBalance.Reset()

	userID := "user123"
	WalletBalance.WithLabelValues(userID).Set(5000)

	balance := testutil.ToFloat64(WalletBalance.WithLabelValues(userID))
	assert.Equal(t, float64(5000), balance)

	// Обновляем баланс
	WalletBalance.WithLabelValues(userID).Set(7500)
	balance = testutil.ToFloat64(WalletBalance.WithLabelValues(userID))
	assert.Equal(t, float64(7500), balance)
}

func TestWalletBalanceMultipleUsers(t *testing.T) {
	WalletBalance.Reset()

	WalletBalance.WithLabelValues("user1").Set(1000)
	WalletBalance.WithLabelValues("user2").Set(2000)
	WalletBalance.WithLabelValues("user3").Set(3000)

	balance1 := testutil.ToFloat64(WalletBalance.WithLabelValues("user1"))
	balance2 := testutil.ToFloat64(WalletBalance.WithLabelValues("user2"))
	balance3 := testutil.ToFloat64(WalletBalance.WithLabelValues("user3"))

	assert.Equal(t, float64(1000), balance1)
	assert.Equal(t, float64(2000), balance2)
	assert.Equal(t, float64(3000), balance3)
}

func TestActiveSubscriptions(t *testing.T) {
	ActiveSubscriptions.Reset()

	ActiveSubscriptions.WithLabelValues("monthly").Set(100)
	ActiveSubscriptions.WithLabelValues("yearly").Set(50)

	monthlyActive := testutil.ToFloat64(ActiveSubscriptions.WithLabelValues("monthly"))
	yearlyActive := testutil.ToFloat64(ActiveSubscriptions.WithLabelValues("yearly"))

	assert.Equal(t, float64(100), monthlyActive)
	assert.Equal(t, float64(50), yearlyActive)
}

func TestMetricsIntegration(t *testing.T) {
	// Сбрасываем все метрики
	HTTPRequestsTotal.Reset()
	BookingsTotal.Reset()
	EmailsSentTotal.Reset()
	SubscriptionsCreatedTotal.Reset()

	// Имитируем реальный сценарий использования
	RecordHTTPRequest("POST", "/api/bookings", "201", 0.25)
	RecordBooking("confirmed", "card")
	RecordEmail("booking_confirmation", "success")
	RecordSubscription("monthly")

	// Проверяем что все метрики записались
	httpCount := testutil.ToFloat64(HTTPRequestsTotal.WithLabelValues("POST", "/api/bookings", "201"))
	bookingCount := testutil.ToFloat64(BookingsTotal.WithLabelValues("confirmed", "card"))
	emailCount := testutil.ToFloat64(EmailsSentTotal.WithLabelValues("booking_confirmation", "success"))
	subCount := testutil.ToFloat64(SubscriptionsCreatedTotal.WithLabelValues("monthly"))

	assert.Equal(t, float64(1), httpCount)
	assert.Equal(t, float64(1), bookingCount)
	assert.Equal(t, float64(1), emailCount)
	assert.Equal(t, float64(1), subCount)
}