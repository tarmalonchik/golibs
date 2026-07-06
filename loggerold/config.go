package loggerold

type Config struct {
	TgBotToken    string `env:"TELEGRAM_BOT_TOKEN" required:"true"`
	TgBotBaseURL  string `env:"TELEGRAM_BOT_API_BASE_URL" required:"true"`
	SelfIPAddress string `env:"SELF_IP_ADDRESS" envDefault:""`
	IsCronJob     bool   `env:"IS_CRON_JOB" required:"true"`
	Environment   string `env:"ENVIRONMENT" envDefault:"prod"`
}
