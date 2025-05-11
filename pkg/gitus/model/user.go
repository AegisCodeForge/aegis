package model

type GitusUserStatus int

const (
	NORMAL_USER GitusUserStatus = 1
	NORMAL_USER_APPROVAL_NEEDED GitusUserStatus = 2
	ADMIN GitusUserStatus = 3
	SUPER_ADMIN GitusUserStatus = 4
	DELETED GitusUserStatus = 5
	BANNED GitusUserStatus = 6
)

type GitusUser struct {
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
	Status GitusUserStatus `json: "status"`
	// AuthKey []GitusAuthKey `json:"authKey"`
	// SigningKey []GitusSigningKey `json:"signKey"`
}

