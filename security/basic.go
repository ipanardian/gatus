package security

import "errors"

var ErrInvalidBasicConfig = errors.New("invalid basic authentication configuration")

// BasicConfig is the configuration for Basic authentication
type BasicConfig struct {
	// Username is the name which will need to be used for a successful authentication
	Username string `yaml:"username"`

	// PasswordBcryptHashBase64Encoded is the base64 encoded string of the Bcrypt hash of the password to use to
	// authenticate using basic auth.
	PasswordBcryptHashBase64Encoded string `yaml:"password-bcrypt-base64"`
}

// IsValid reports whether the basic security configuration is complete.
func (c *BasicConfig) IsValid() bool {
	return c != nil && len(c.Username) > 0 && len(c.PasswordBcryptHashBase64Encoded) > 0
}

// isValid is retained for existing package callers.
func (c *BasicConfig) isValid() bool {
	return c.IsValid()
}
