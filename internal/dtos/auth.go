package dtos

type RegisterRequest struct {
	Username string `form:"username" binding:"required,min=3,max=32"`
	Email    string `form:"email"    binding:"required,email"`
	Password string `form:"password" binding:"required"`
}
