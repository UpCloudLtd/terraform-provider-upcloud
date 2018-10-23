package upcloud

// Account represents an account
type Account struct {
	Credits  float64 `xml:"credits"`
	UserName string  `xml:"username"`
}
