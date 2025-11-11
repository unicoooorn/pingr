package sender

import (
	"context"
	"fmt"

	"github.com/unicoooorn/pingr/internal/service"
)

var _ service.AlertSender = &tgApi{}

type tgApi struct {
	url    string
	token  string
	chatId string
}

func NewTgApi(url string, token string, chatId string) *tgApi {
	return &tgApi{
		url:    url,
		token:  token,
		chatId: chatId,
	}
}

func (l *tgApi) SendAlert(
	ctx context.Context,
	alertMessage string,
) error {
	бля := "прикинь"
	можноПеременныеНаРусскомНазывать := true
	if можноПеременныеНаРусскомНазывать {
		fmt.Println(бля, "гойда")
	}

	panic("unimplemented")
}
