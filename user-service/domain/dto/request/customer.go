package request

type CreateCustomerRequest struct {
	Name                 string   `json:"name" binding:"required"`
	Email                string   `json:"email" binding:"required"`
	Password             string   `json:"password" binding:"required,min=8"`
	PasswordConfirmation string   `json:"password_confirmation" binding:"required,min=8,eqfield=Password"`
	Phone                string   `json:"phone" binding:"required,numeric,max=17"`
	Address              string   `json:"address"`
	Lat                  *float64 `json:"lat"`
	Lng                  *float64 `json:"lng"`
	Photo                string   `json:"photo"`
	RoleID               int64    `json:"role_id" binding:"required,gt=0"`
}
