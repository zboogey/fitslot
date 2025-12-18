package server

import (
	"context"
	"fitslot/internal/auth"
	"fitslot/internal/booking"
	"fitslot/internal/config"
	"fitslot/internal/email"
	"fitslot/internal/gym"
	"fitslot/internal/subscription"
	"fitslot/internal/user"
	"fitslot/internal/wallet"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type Server struct {
	router     *gin.Engine
	db         *sqlx.DB
	config     *config.Config
	email      *email.Service
	httpServer *http.Server
}

func New(db *sqlx.DB, cfg *config.Config, emailService *email.Service) *Server {
	router := gin.Default()
	router.Use(MetricsMiddleware())
	router.Use(RequestLoggingMiddleware())
	router.Use(RateLimitMiddleware(100, 200)) // 100 requests per second, burst of 200
	router.Use(corsMiddleware())

	userRepo := user.NewRepository(db)
	gymRepo := gym.NewRepository(db)
	bookingRepo := booking.NewRepository(db)
	walletRepo := wallet.NewRepository(db)
	subscriptionRepo := subscription.NewRepository(db)

	userService := user.NewService(userRepo, cfg.JWTSecret)
	gymService := gym.NewService(gymRepo)
	bookingService := booking.NewService(
		bookingRepo,
		gymRepo,
		subscriptionRepo,
		walletRepo,
		userRepo,
		emailService,
	)

	userHandler := user.NewHandler(userService, cfg.JWTSecret)
	gymHandler := gym.NewHandler(gymService)
	bookingHandler := booking.NewHandler(bookingService)
	walletHandler := wallet.NewHandler(walletRepo)
	subscriptionHandler := subscription.NewHandler(subscriptionRepo, walletRepo)
	router.GET("/metrics", Metrics())

	public := router.Group("/auth")
	{
		public.POST("/register", userHandler.Register)
		public.POST("/login", userHandler.Login)
		public.POST("/refresh", userHandler.RefreshToken)
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
		protected.GET("/wallet", walletHandler.GetBalance)
		protected.POST("/wallet/topup", walletHandler.TopUp)
		protected.GET("/wallet/transactions", walletHandler.ListTransactions)
		protected.POST("/subscriptions", subscriptionHandler.Create)
		protected.GET("/subscriptions", subscriptionHandler.ListMy)
		protected.GET("/subscriptions/plans", subscriptionHandler.ListPlans)
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

	SetupSwagger(router)

	router.GET("/health", Health)
	router.GET("/test-email", TestEmail(emailService))

	return &Server{
		router: router,
		db:     db,
		config: cfg,
		email:  emailService,
	}
}

func (s *Server) Start(port string) error {
	addr := ":" + port
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
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
