package api_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hsmtkk/j-quants-api-go/api"
	"github.com/stretchr/testify/assert"
)

func Test0(t *testing.T) {
	mailAddress := os.Getenv("MAIL_ADDRESS")
	password := os.Getenv("PASSWORD")
	clt, err := api.New(mailAddress, password)
	assert.Nil(t, err)
	now := time.Now()
	from := now.AddDate(0, 0, -7)
	to := now.AddDate(0, 0, 7)
	param := api.TradingCalendarParam{
		From: &from,
		To:   &to,
	}
	result, err := clt.TradingCalendar(param)
	assert.Nil(t, err)
	fmt.Printf("%v\n", result)
}
