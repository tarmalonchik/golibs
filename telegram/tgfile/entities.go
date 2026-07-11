package tgfile

const (
	tokenTemp     = "/bot%s"
	documentField = "document"
	chatIDField   = "chat_id"
)

type Config struct {
	TgBotToken   string `env:"TELEGRAM_BOT_TOKEN,required"`
	TgBotBaseURL string `env:"TELEGRAM_BOT_API_BASE_URL,required"`
}

type sendDocResponseBody struct {
	Ok     bool   `json:"ok"`
	Result result `json:"result"`
}

type result struct {
	Document document `json:"document"`
}

type document struct {
	FileID string `json:"file_id"`
}
