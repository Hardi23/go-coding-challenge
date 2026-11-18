package cors_middleware

type Config struct {
	CorsMaxAge           int      `env:"MAX_AGE, default=3600"`
	CorsAllowOriginRegex []string `env:"ALLOWED_ORIGIN_REGEX, default=(.+)"`
	CorsAllowMethods     string   `env:"ALLOW_METHODS, default=GET,POST,OPTIONS,DELETE,PUT"`
	CorsAllowHeaders     string   `env:"ALLOW_HEADERS, default=Authorization,Content-Type"`
	CorsAllowCredentials bool     `env:"ALLOW_CREDENTIALS"`
}
