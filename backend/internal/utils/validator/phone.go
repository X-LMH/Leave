package validator

import "regexp"

var mainlandMobileRegexp = regexp.MustCompile(`^1[3-9][0-9]{9}$`)

// IsMainlandMobile 判断手机号是否为 11 位中国大陆手机号。
func IsMainlandMobile(phone string) bool {
	return mainlandMobileRegexp.MatchString(phone)
}
