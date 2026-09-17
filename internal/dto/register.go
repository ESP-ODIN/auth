package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,alphanum,min=3,max=30"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}
