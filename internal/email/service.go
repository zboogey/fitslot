package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/redis/go-redis/v9"
	"fitslot/internal/logger"
)

type EmailJob struct {
	To      string    `json:"to"`
	Name    string    `json:"name"`
	Subject string    `json:"subject"`
	Body    string    `json:"body"`
	Tries   int       `json:"tries"`
	Created time.Time `json:"created"`
}

type Service struct {
	redis    *redis.Client
	from     string
	fromName string
	smtpHost string
	smtpPort string
	smtpUser string
	smtpPass string
}

func New(fromEmail, fromName, smtpHost, smtpPort, smtpUser, smtpPass, redisAddr string) *Service {
	return &Service{
		redis: redis.NewClient(&redis.Options{
			Addr: redisAddr,
		}),
		from:     fromEmail,
		fromName: fromName,
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		smtpUser: smtpUser,
		smtpPass: smtpPass,
	}
}

func (s *Service) Send(ctx context.Context, to, name, subject, body string) error {
	job := EmailJob{
		To:      to,
		Name:    name,
		Subject: subject,
		Body:    body,
		Tries:   0,
		Created: time.Now(),
	}

	data, err := json.Marshal(job)
	if err != nil {
		logger.Errorf("Failed to marshal email job: %v", err)
		return err
	}

	if err := s.redis.LPush(ctx, "emails", data).Err(); err != nil {
		logger.Errorf("Failed to queue email to %s: %v", to, err)
		return err
	}

	logger.Infof("Email queued: %s to %s", subject, to)
	return nil
}

func (s *Service) Start(ctx context.Context) {
	logger.Info("Email service started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Email service stopped")
			return
		default:
			s.processNext(ctx)
		}
	}
}

func (s *Service) processNext(ctx context.Context) {
	result, err := s.redis.BRPop(ctx, 2*time.Second, "emails").Result()
	if err != nil {
		return
	}

	var job EmailJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		logger.Errorf("Bad email data: %v", err)
		return
	}

	job.Tries++
	logger.Infof("Sending email to %s (attempt %d)", job.To, job.Tries)
	if err := s.sendNow(job); err != nil {
		logger.Errorf("Failed to send email to %s: %v", job.To, err)

		if job.Tries < 3 {
			time.Sleep(5 * time.Second)
			data, _ := json.Marshal(job)
			s.redis.LPush(context.Background(), "emails", data)
			logger.Infof("Retrying email to %s (attempt %d)", job.To, job.Tries+1)
		} else {
			logger.Errorf("Email to %s failed after 3 attempts", job.To)
			s.saveFailed(job, err)
		}
		return
	}

	logger.Infof("Email sent successfully to %s", job.To)
}

func (s *Service) sendNow(job EmailJob) error {

	message := fmt.Sprintf("From: %s <%s>\r\n", s.fromName, s.from)
	message += fmt.Sprintf("To: %s\r\n", job.To)
	message += fmt.Sprintf("Subject: %s\r\n", job.Subject)
	message += "\r\n" + job.Body

	var auth smtp.Auth
	if s.smtpUser != "" && s.smtpPass != "" {
		auth = smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)
	}

	addr := s.smtpHost + ":" + s.smtpPort
	err := smtp.SendMail(addr, auth, s.from, []string{job.To}, []byte(message))

	return err
}

func (s *Service) saveFailed(job EmailJob, err error) {
	failed := map[string]interface{}{
		"job":   job,
		"error": err.Error(),
		"time":  time.Now(),
	}
	data, _ := json.Marshal(failed)
	s.redis.LPush(context.Background(), "emails:failed", data)
	logger.Errorf("Email moved to failed queue: %s", job.To)
}

func (s *Service) QueueLength(ctx context.Context) int64 {
	length, _ := s.redis.LLen(ctx, "emails").Result()
	return length
}

func (s *Service) Close() error {
	return s.redis.Close()
}

func (s *Service) SendBookingConfirmation(ctx context.Context, email, name, bookingType, details string, when time.Time) error {
	subject := "Booking Confirmed - " + bookingType
	body := fmt.Sprintf(`Hi %s,

Your booking is confirmed!

Type: %s
Details: %s
Time: %s

See you at the gym!

- FitSlot Team`, name, bookingType, details, when.Format("Jan 2, 2006 at 3:04 PM"))

	return s.Send(ctx, email, name, subject, body)
}

func (s *Service) SendReminder(ctx context.Context, email, name, bookingType, details string, when time.Time) error {
	subject := "Reminder: " + bookingType + " Tomorrow"
	body := fmt.Sprintf(`Hi %s,

This is a reminder about your booking tomorrow:

Type: %s
Details: %s
Time: %s

See you soon!

- FitSlot Team`, name, bookingType, details, when.Format("Jan 2, 2006 at 3:04 PM"))

	return s.Send(ctx, email, name, subject, body)
}

func (s *Service) SendCancellation(ctx context.Context, email, name, bookingType, details string) error {
	subject := "Booking Cancelled - " + bookingType
	body := fmt.Sprintf(`Hi %s,

Your booking has been cancelled:

Type: %s
Details: %s


- FitSlot Team`, name, bookingType, details)

	return s.Send(ctx, email, name, subject, body)
}
