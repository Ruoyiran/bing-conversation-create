package bing

type Author string

type ConversationResult struct {
	ConversationId        string
	ClientId              string
	ConversationSignature string
	Result                APIResult
}

type APIResult struct {
	Value   string
	Message interface{}
}
