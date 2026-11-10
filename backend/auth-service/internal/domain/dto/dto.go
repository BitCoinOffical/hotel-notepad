package dto

type RegisterDTO struct {
	UserName      string `json:"username" binding:"min=2,max=16,required"`
	Password      string `json:"password" binding:"min=8,required"`
	PasswordRetry string `json:"password_retry" binding:"min=8,required"`
}
