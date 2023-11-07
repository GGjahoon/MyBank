package val

import (
	"fmt"
	"net/mail"
	"regexp"
)

var (
	// 正则表达式： ``中 ^:字符串的开始 []:可能出现的字符为：a-z、0-9和下划线 +:可能出现的字符可以出现多次 $:字符串结尾
	// \\s:表示任何空格字符
	isValidUsername = regexp.MustCompilePOSIX(`^[a-z0-9_]+$`).MatchString
	isValidFullName = regexp.MustCompilePOSIX(`[a-zA-Z\\\\s]+$`).MatchString
)

func ValidateString(value string, minLength int, maxLength int) error {
	n := len(value)
	if n < minLength || n > maxLength {
		return fmt.Errorf("must contain from %d - %d characters ", minLength, maxLength)
	}
	return nil
}
func ValidateUsername(value string) error {
	err := ValidateString(value, 3, 100)
	if err != nil {
		return err
	}
	if !isValidUsername(value) {
		return fmt.Errorf("must contain only lower letters,digits,or under score")
	}
	return nil
}
func ValidateFullName(value string) error {
	err := ValidateString(value, 3, 100)
	if err != nil {
		return err
	}
	if !isValidFullName(value) {
		return fmt.Errorf("must contain only letters or spaces")
	}
	return nil
}
func ValidatePassword(value string) error {
	return ValidateString(value, 6, 100)
}
func ValidateEmail(value string) error {
	if err := ValidateString(value, 3, 200); err != nil {
		return err
	}
	if _, err := mail.ParseAddress(value); err != nil {
		return fmt.Errorf("is not a valid email address")
	}
	return nil
}
func ValidateEmailID(value int64) error {
	if value <= 0 {
		return fmt.Errorf("must be a postive integer")
	}
	return nil
}
func ValidateSecret(value string) error {
	return ValidateString(value, 32, 128)
}
