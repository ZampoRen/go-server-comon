package cachekey

const (
	UserInfoKey = "USER_INFO:"
)

func GetUserInfoKey(userID string) string {
	return UserInfoKey + userID
}
