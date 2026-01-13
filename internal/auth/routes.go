package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *AuthHandler, jwtManager *JWTManager) {
	public := r.Group("/auth")
	{
		public.POST("/register", h.StartRegistration)
		public.POST("/register-confirm", h.CreateNewUser)
		public.POST("/login", h.Login)
	}

	protected := r.Group("/auth")
	protected.Use(AuthMiddleware(jwtManager))
	{
		protected.POST("/refresh", h.Refresh)
		protected.POST("/logout", h.Logout)
	}
}
