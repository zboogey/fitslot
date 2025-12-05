package server

import (
	"fitslot/internal/auth"
	"fitslot/internal/booking"
	"fitslot/internal/config"
	"fitslot/internal/email"
	"fitslot/internal/gym"
	"fitslot/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	router *gin.Engine
	db     *sqlx.DB
	config *config.Config
	email  *email.Service
}

func New(db *sqlx.DB, cfg *config.Config, emailService *email.Service) *Server {
	router := gin.Default()
	router.Use(corsMiddleware())

	userHandler := user.NewHandler(db, cfg.JWTSecret)
	gymHandler := gym.NewHandler(db)
	bookingHandler := booking.NewHandler(db, emailService)

	public := router.Group("/auth")
	{
		public.POST("/register", userHandler.Register)
		public.POST("/login", userHandler.Login)
	}

	authMiddleware := auth.AuthMiddleware(cfg.JWTSecret)
	protected := router.Group("/")
	protected.Use(authMiddleware)
	{
		protected.GET("/me", userHandler.GetMe)
		protected.GET("/gyms", gymHandler.ListGyms)
		protected.GET("/gyms/:gymID/slots", gymHandler.ListTimeSlots)
		protected.POST("/slots/:slotID/book", bookingHandler.BookSlot)
		protected.POST("/bookings/:bookingID/cancel", bookingHandler.CancelBooking)
		protected.GET("/bookings", bookingHandler.ListMyBookings)
	}

	adminMiddleware := auth.RequireRole("admin")
	admin := router.Group("/admin")
	admin.Use(authMiddleware, adminMiddleware)
	{
		admin.POST("/gyms", gymHandler.CreateGym)
		admin.GET("/gyms", gymHandler.ListGyms)
		admin.POST("/gyms/:gymID/slots", gymHandler.CreateTimeSlot)
		admin.GET("/gyms/:gymID/slots", gymHandler.ListTimeSlots)
		admin.GET("/slots/:slotID/bookings", bookingHandler.ListBookingsBySlot)
		admin.GET("/gyms/:gymID/bookings", bookingHandler.ListBookingsByGym)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	
	router.GET("/test-email", func(c *gin.Context) {
		testEmail := c.Query("email")
		if testEmail == "" {
			c.JSON(400, gin.H{"error": "email parameter required"})
			return
		}
		
		err := emailService.Send(c.Request.Context(), testEmail, "Test User", "Test Email", "Email service is working!")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		
		c.JSON(200, gin.H{"status": "Email queued successfully"})
	})

	return &Server{
		router: router,
		db:     db,
		config: cfg,
		email:  emailService,
	}
}

func (s *Server) Start(port string) error {
	addr := ":" + port
	return s.router.Run(addr)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
