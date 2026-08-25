package domain

type UserRegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserLoginRequest UserRegisterRequest

type UserLoginResponse struct {
	Token string `json:"token"`
}

type UserRegisterResponse UserLoginResponse

type CreateUserParams struct {
	Login        string
	PasswordHash string
	Salt         []byte
}

type User struct {
	ID           int64
	Login        string
	PasswordHash string
}
