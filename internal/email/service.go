package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
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
	sendgrid *sendgrid.Client
	redis    *redis.Client
	from     string
	fromName string
}

func New(sendgridKey, fromEmail, fromName, redisAddr string) *Service {
	return &Service{
		sendgrid: sendgrid.NewSendClient(sendgridKey),
		redis: redis.NewClient(&redis.Options{
			Addr: redisAddr,
		}),
		from:     fromEmail,
		fromName: fromName,
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
		return err
	}

	return s.redis.LPush(ctx, "emails", data).Err()
}

func (s *Service) Start(ctx context.Context) {
	log.Println("Email service started")
	
	for {
		select {
		case <-ctx.Done():
			log.Println("Email service stopped")
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
		log.Printf("Bad email data: %v", err)
		return
	}

	job.Tries++
	log.Printf("Sending email to %s (attempt %d)", job.To, job.Tries)

	if err := s.sendNow(job); err != nil {
		log.Printf("Failed to send email: %v", err)

		if job.Tries < 3 {
			time.Sleep(5 * time.Second)
			data, _ := json.Marshal(job)
			s.redis.LPush(context.Background(), "emails", data)
			log.Printf("Will retry email to %s", job.To)
		} else {
			log.Printf("Email to %s failed after 3 tries", job.To)
			s.saveFailed(job, err)
		}
		return
	}

	log.Printf("Email sent to %s", job.To)
}

func (s *Service) sendNow(job EmailJob) error {
	from := mail.NewEmail(s.fromName, s.from)
	to := mail.NewEmail(job.Name, job.To)
	message := mail.NewSingleEmail(from, job.Subject, to, job.Body, job.Body)

	response, err := s.sendgrid.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error: %d", response.StatusCode)
	}

	return nil
}

func (s *Service) saveFailed(job EmailJob, err error) {
	failed := map[string]interface{}{
		"job":   job,
		"error": err.Error(),
		"time":  time.Now(),
	}
	data, _ := json.Marshal(failed)
	s.redis.LPush(context.Background(), "emails:failed", data)
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