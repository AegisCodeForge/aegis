package model

type AegisUserStatus int

const (
	NORMAL_USER AegisUserStatus = 1
	NORMAL_USER_APPROVAL_NEEDED AegisUserStatus = 2
	NORMAL_USER_CONFIRM_NEEDED AegisUserStatus = 3
	ADMIN AegisUserStatus = 4
	SUPER_ADMIN AegisUserStatus = 5
	BANNED AegisUserStatus = 7
)

type AegisUser struct {
	// user name. 
	Name string `json:"name"`
	// user "title"
	Title string `json:"title"`
	// user email.
	Email string `json:"email"`
	// user bio.
	Bio string `json:"bio"`
	Website string `json:"website"`
	// password hash.
	PasswordHash string `json:"passwordHash"`
	RegisterTime int64 `json:"regTime"`
	Status AegisUserStatus `json: "status"`
	// AuthKey []AegisAuthKey `json:"authKey"`
	// SigningKey []AegisSigningKey `json:"signKey"`
}

