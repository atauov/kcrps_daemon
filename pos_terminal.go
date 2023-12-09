package daemon

type PosTerminal struct {
	Id         string `db:"pos_id"`
	UserId     int    `db:"user_id"`
	WebHookURL string `db:"webhook_url"`
	FlaskId    int    `db:"flask_id"`
}
