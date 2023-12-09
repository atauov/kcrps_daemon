package daemon

import "github.com/google/uuid"

type PosTerminal struct {
	Id         uuid.UUID `db:"pos_id"`
	UserId     int       `db:"user_id"`
	WebHookURL string    `db:"webhook_url"`
	FlaskId    int       `db:"flask_id"`
}
