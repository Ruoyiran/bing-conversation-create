package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Ruoyiran/endless"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"log"
	"openai-api-proxy/bing"
)

var (
	port = 8080
	host = "0.0.0.0"
)

func init() {
	flag.StringVar(&host, "host", "0.0.0.0", "server host")
	flag.IntVar(&port, "port", 8080, "server port")
}

func main() {
	flag.Parse()

	handler := gin.Default()
	handler.POST("/bing/conversation/create", createBingConversation)

	gin.SetMode(gin.ReleaseMode)

	logrus.Infof("[*] Starting server at %s:%d\n", host, port)

	err := endless.ListenAndServe(fmt.Sprintf("%s:%d", host, port), handler)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func createBingConversation(c *gin.Context) {
	type reqParam struct {
		Cookie string `json:"cookie" binding:"required"`
	}
	req := reqParam{}
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.Errorf("%s", err.Error())
		sendFailResponse(c, errors.New("cookie is required"))
		return
	}
	conversation, err := bing.CreateConversationWithRetry(req.Cookie, 3)
	if err != nil {
		logrus.Errorf("CreateConversation failed, error: %s", err.Error())
		sendFailResponse(c, err)
	} else {
		logrus.Infof("CreateConversation succeed, clientId: %s, conversationId: %s, conversationSignature: %s",
			conversation.ClientId, conversation.ConversationId, conversation.ConversationSignature)
		sendSuccessResponse(c, conversation)
	}
}
