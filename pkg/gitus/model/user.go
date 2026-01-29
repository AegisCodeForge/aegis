package model

type GitusUserStatus int

const (
	NORMAL_USER GitusUserStatus = 1
	NORMAL_USER_APPROVAL_NEEDED GitusUserStatus = 2
	NORMAL_USER_CONFIRM_NEEDED GitusUserStatus = 3
	ADMIN GitusUserStatus = 4
	SUPER_ADMIN GitusUserStatus = 5
	BANNED GitusUserStatus = 7
	NORMAL_USER_NO_NEW_NAMESPACE GitusUserStatus = 8
)

func ValidUserName(s string) bool {
	for _, k := range s {
		if !(('0' <= k && k <= '9') || ('A' <= k && k <= 'Z') || ('a' <= k && k <= 'z') || k == '_' || k == '-') { return false }
	}
	return true
}

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
	Status GitusUserStatus `json:"status"`
	// AuthKey []GitusAuthKey `json:"authKey"`
	// SigningKey []GitusSigningKey `json:"signKey"`

	TFAConfig GitusUser2FAConfig `json:"2fa"`
	WebsitePreference GitusUserWebsitePreference `json:"preference"`
}

type GitusUser2FAConfig struct {
	Email struct{
		Enable bool `json:"enable"`
	} `json:"email"`
}

type GitusUserWebsitePreference struct {
	ForegroundColor string `json:"foregroundColor"`
	BackgroundColor string `json:"backgroundColor"`
	// ignore user custom fgcolor/bgcolor config if true.
	UseSiteWideThemeConfig bool `json:"useSiteWideThemeConfig"`
	// whether to load ui w/ components that requires javascript or
	// load ui with zero javascript requirements.
	UseJavascript bool `json:"useJavascript"`
}



