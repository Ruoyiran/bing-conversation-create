package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

type ResponseWithResult struct {
	Response
	Result interface{} `json:"result"`
}

func getResponse(c *gin.Context, err error) *Response {
	if err == nil {
		return &Response{
			Code: 0,
			Msg:  "Success",
		}
	} else {
		return &Response{
			Code: -1,
			Msg:  err.Error(),
		}
	}
}

func getResponseWithResult(c *gin.Context, err error, result interface{}) *ResponseWithResult {
	if err == nil {
		return &ResponseWithResult{
			Response: Response{
				Code: 0,
				Msg:  "Success",
			},
			Result: result,
		}
	} else {
		return &ResponseWithResult{
			Response: Response{
				Code: -1,
				Msg:  err.Error(),
			},
			Result: result,
		}
	}
}

func sendSuccessResponse(c *gin.Context, result interface{}) {
	sendResponse(c, nil, result)
}

func sendFailResponse(c *gin.Context, err error) {
	sendResponse(c, err, nil)
}

func sendResponse(c *gin.Context, err error, result interface{}) {
	if result == nil {
		c.JSON(http.StatusOK, getResponse(c, err))
	} else {
		c.JSON(http.StatusOK, getResponseWithResult(c, err, result))
	}
}
