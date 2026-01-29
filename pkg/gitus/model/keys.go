package model

type GitusSigningKey struct {
	UserName string `json:"userName"`
	KeyName string `json:"keyName"`
	KeyText string `json:"keyText"`
}

type GitusAuthKey struct {
	UserName string `json:"userName"`
	KeyName string `json:"keyName"`
	KeyText string `json:"keyText"`
}

