package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/bearathome/gologger"
	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/linebot"
)

const ACCESS_KEY = "LINE_ACCESS"
const SECRET_KEY = "LINE_SECRET"
const CHANNEL_KEY = "LINE_CHANNEL"

func getEnv() (string, string) {
	return os.Getenv(SECRET_KEY), os.Getenv(ACCESS_KEY)
}

func main() {
	secret, access := getEnv()
	gologger.SetLogLevel(gologger.LogLevelDebug)
	gologger.Debug("Load env with %s, %s", secret, access)

	bot, err := linebot.New(secret, access)
	if err != nil {
		gologger.Error("linebot init fail %s", err.Error())
		os.Exit(1)
	}

	r := gin.Default()
	r.POST("/", func(c *gin.Context) {
		events, err := bot.ParseRequest(c.Request)
		if err != nil {
			gologger.Error("Parse line req fail: %s", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		for idx, event := range events {
			err = handleEvent(bot, event)
			if err != nil {
				gologger.Error("Parse event %d fail: %s", idx, err.Error())
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"error": "",
		})
	})
	r.Run("0.0.0.0:5221")
}

func handleEvent(bot *linebot.Client, event *linebot.Event) error {
	switch event.Type {
	case linebot.EventTypeMessage:
		return handleMessage(bot, event)
	}
	return errors.New("unsupported event type")
}

func handleMessage(bot *linebot.Client, event *linebot.Event) error {
	switch t := event.Message.(type) {
	case *linebot.TextMessage:
		msg := event.Message.(*linebot.TextMessage)
		return replyText(bot, event.ReplyToken, fmt.Sprintf("Reply: %s", msg.Text))
	case *linebot.ImageMessage:
		msg := event.Message.(*linebot.ImageMessage)
		return replyText(bot, event.ReplyToken,
			fmt.Sprintf("Reply: %s\nPreview: %s", msg.OriginalContentURL, msg.PreviewImageURL))
	default:
		return fmt.Errorf("unsupported message type: %+v", t)
	}
}

func replyText(bot *linebot.Client, replyToken, text string) error {
	reply := linebot.NewTextMessage(text)
	_, err := bot.ReplyMessage(replyToken, reply).Do()
	if err != nil {
		gologger.Error("reply fail: %s", err.Error())
	}
	return err
}
