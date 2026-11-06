package dto

type RegisterRequest struct {
	Email       string `json:"email"         binding:"required,email"                    example:"jane.doe@example.com"`
	Password    string `json:"password"      binding:"required,min=8,max=100"            example:"Str0ngPassw0rd!"`
	Username    string `json:"username"      binding:"required,min=3,max=50"             example:"janedoe"`
	FirstName   string `json:"first_name"    binding:"required,min=3,max=100"            example:"Jane"`
	LastName    string `json:"last_name"     binding:"required,min=2,max=100"            example:"Doe"`
	DateOfBirth string `json:"date_of_birth" binding:"required,datetime=2006-01-02" example:"1990-01-31" format:"date"`
}
