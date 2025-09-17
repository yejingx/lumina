package dao

type RegisterRequest struct {
	AccessToken string `json:"accessToken" binding:"required"`
	Uuid        string `json:"uuid"`
}

type S3Config struct {
	AccessKeyID     *string `json:"accessKeyID"`
	SecretAccessKey *string `json:"secretAccessKey"`
}

type RegisterResponse struct {
	Uuid              string  `json:"uuid"`
	Token             string  `json:"token"`
	S3AccessKeyID     *string `json:"s3AccessKeyID"`
	S3SecretAccessKey *string `json:"s3SecretAccessKey"`
}