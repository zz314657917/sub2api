package service

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

func requiresSystemChatRole(account *Account, targetURL string) bool {
	if account == nil || account.Type != AccountTypeAPIKey {
		return false
	}
	switch account.Platform {
	case PlatformDeepseek, PlatformKimi, PlatformZhipu:
		return true
	}

	// Compatible accounts require an exact selected upstream host. Model names
	// and path/userinfo fragments must not influence this transport adaptation.
	u, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	switch strings.ToLower(u.Hostname()) {
	case "api.deepseek.com", "api.kimi.com", "api.moonshot.cn", "api.moonshot.ai", "open.bigmodel.cn", "api.z.ai":
		return true
	}
	return false
}

// normalizeStrictChatDeveloperRoles adapts native Chat requests after account
// selection. RawMessage keeps unknown fields and numeric literals intact, and
// returns the original input when no adaptation is needed for a retry.
func normalizeStrictChatDeveloperRoles(account *Account, targetURL string, body []byte) ([]byte, error) {
	if !requiresSystemChatRole(account, targetURL) {
		return body, nil
	}

	var root map[string]json.RawMessage
	if json.Unmarshal(body, &root) != nil || root == nil {
		return nil, errors.New("chat role normalization: invalid request object")
	}
	rawMessages, exists := root["messages"]
	if !exists {
		return body, nil
	}
	var messages []json.RawMessage
	if json.Unmarshal(rawMessages, &messages) != nil {
		return nil, errors.New("chat role normalization: invalid messages array")
	}

	changed := false
	for i, rawMessage := range messages {
		var message map[string]json.RawMessage
		if json.Unmarshal(rawMessage, &message) != nil || message == nil {
			return nil, errors.New("chat role normalization: invalid message object")
		}
		var role string
		if json.Unmarshal(message["role"], &role) != nil {
			return nil, errors.New("chat role normalization: invalid message role")
		}
		if role != "developer" {
			continue
		}
		message["role"] = json.RawMessage(`"system"`)
		updated, err := json.Marshal(message)
		if err != nil {
			return nil, errors.New("chat role normalization: cannot encode message")
		}
		messages[i] = updated
		changed = true
	}
	if !changed {
		return body, nil
	}
	updatedMessages, err := json.Marshal(messages)
	if err != nil {
		return nil, errors.New("chat role normalization: cannot encode messages")
	}
	root["messages"] = updatedMessages
	return json.Marshal(root)
}
