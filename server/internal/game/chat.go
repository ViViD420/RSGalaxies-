package game

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
	"time"
)

const (
	MaxChatMessageRunes = 240
	ChatCooldown = 2 * time.Second
)

var (
	ErrInvalidChatMessage = errors.New("chat message must contain 1-240 characters")
	ErrChatCooldown = errors.New("chat is cooling down")
	ErrInvalidChatCursor = errors.New("invalid chat cursor or page size")
)

type ChatMessage struct {
	ID int64 `json:"id"`
	ZoneID string `json:"zoneId"`
	CharacterID PlayerID `json:"characterId"`
	CharacterName string `json:"characterName"`
	Message string `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

func ValidateChatMessage(message string)(string,error){
	message=strings.Map(func(r rune)rune{
		if r=='\t'{return ' '}
		if unicode.IsControl(r){return -1}
		return r
	},message)
	message=strings.TrimSpace(message)
	length:=utf8.RuneCountInString(message)
	if length<1||length>MaxChatMessageRunes{return "",ErrInvalidChatMessage}
	return message,nil
}
