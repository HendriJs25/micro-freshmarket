package request

type VerifyAccountQuery struct {
	Token string `form:"token" binding:"required"`
}
