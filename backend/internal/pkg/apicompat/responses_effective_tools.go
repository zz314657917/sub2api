package apicompat

import "encoding/json"

// EffectiveResponsesTools includes completed client discoveries without changing the request.
func EffectiveResponsesTools(req *ResponsesRequest) ([]ResponsesTool, error) {
	tools := append([]ResponsesTool(nil), req.Tools...)
	input := bytesTrimSpace(req.Input)
	if len(input) == 0 || input[0] != '[' {
		return tools, nil
	}
	var rawInput []any
	if err := json.Unmarshal(input, &rawInput); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(tools)
	if err != nil {
		return nil, err
	}
	var rawTools []any
	if err := json.Unmarshal(encoded, &rawTools); err != nil {
		return nil, err
	}
	promoted, err := promotedResponsesToolSearchDiscoveries(rawTools, rawInput)
	if err != nil {
		return nil, err
	}
	encoded, err = json.Marshal(promoted)
	if err != nil {
		return nil, err
	}
	var discovered []ResponsesTool
	if err := json.Unmarshal(encoded, &discovered); err != nil {
		return nil, err
	}
	return append(tools, discovered...), nil
}
