package game

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateChatMessage(t *testing.T) {
	valid,err:=ValidateChatMessage("  Hello frontier!  ")
	if err!=nil||valid!="Hello frontier!"{t.Fatalf("validated message=%q err=%v",valid,err)}
	if _,err:=ValidateChatMessage(" \n\t ");!errors.Is(err,ErrInvalidChatMessage){t.Fatalf("empty message error=%v",err)}
	if _,err:=ValidateChatMessage(strings.Repeat("界",MaxChatMessageRunes+1));!errors.Is(err,ErrInvalidChatMessage){t.Fatalf("oversized message error=%v",err)}
	sanitized,err:=ValidateChatMessage("hello\nworld")
	if err!=nil||sanitized!="helloworld"{t.Fatalf("control character sanitization=%q err=%v",sanitized,err)}
}
