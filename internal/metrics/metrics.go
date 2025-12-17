package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fitslot_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fitslot_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	BookingsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fitslot_bookings_total",
			Help: "Total number of bookings",
		},
		[]string{"status", "payment_method"},
	)

	BookingCancellationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "fitslot_booking_cancellations_total",
			Help: "Total number of booking cancellations",
		},
	)

	EmailsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fitslot_emails_sent_total",
			Help: "Total number of emails sent",
		},
		[]string{"type", "status"},
	)

	EmailQueueLength = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "fitslot_email_queue_length",
			Help: "Current length of email queue",
		},
	)

	WalletTopUpsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "fitslot_wallet_topups_total",
			Help: "Total number of wallet top-ups",
		},
	)

	WalletBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fitslot_wallet_balance_cents",
			Help: "Current wallet balance in cents",
		},
		[]string{"user_id"},
	)

	SubscriptionsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fitslot_subscriptions_created_total",
			Help: "Total number of subscriptions created",
		},
		[]string{"type"},
	)

	ActiveSubscriptions = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fitslot_active_subscriptions",
			Help: "Number of active subscriptions",
		},
		[]string{"type"},
	)
)

func RecordHTTPRequest(method, path, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

func RecordBooking(status, paymentMethod string) {
	BookingsTotal.WithLabelValues(status, paymentMethod).Inc()
}

func RecordBookingCancellation() {
	BookingCancellationsTotal.Inc()
}

func RecordEmail(emailType, status string) {
	EmailsSentTotal.WithLabelValues(emailType, status).Inc()
}

func RecordWalletTopUp() {
	WalletTopUpsTotal.Inc()
}

func RecordSubscription(subType string) {
	SubscriptionsCreatedTotal.WithLabelValues(subType).Inc()
}