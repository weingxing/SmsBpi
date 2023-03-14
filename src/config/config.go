package config

type Config struct {
	Device       string `json:"device"`
	BaudRate     uint   `json:"baudrate"`
	BarkServer   string `json:"barkServer"`
	BarkSecret   string `json:"barkSecret"`
	WxCorpid     string `json:"wxCorpId"`
	WxCorpSecret string `json:"wxCorpSecret"`
	WxAgentid    uint   `json:"wxAgentId"`
	WxUser       string `json:"wxUser"`
	Email        string `json:"email"`
	SmtpServer   string `json:"smtpServer"`
	SmtpPort     uint   `json:"smtpPort"`
	SmtpUser     string `json:"smtpUser"`
	SmtpPassword string `json:"smtpPassword"`
	MqAddress    string `json:"mqAddress"`
	Topic        string `json:"topic"`
}
