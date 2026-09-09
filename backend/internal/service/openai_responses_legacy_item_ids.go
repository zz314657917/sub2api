package service

import (
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// Older protocol bridges emitted item_* IDs for function calls. Drop only
// those optional IDs on replay; preserve call_id and all unrelated JSON bytes.
func stripLegacyResponsesFunctionItemIDs(body []byte) ([]byte, bool, error) {
	input := gjson.GetBytes(body, "input")
	if !input.IsArray() {
		return body, false, nil
	}
	changed := false
	for i, item := range input.Array() {
		if item.Get("type").String() != "function_call" || !strings.HasPrefix(item.Get("id").String(), "item_") {
			continue
		}
		var err error
		body, err = sjson.DeleteBytes(body, "input."+strconv.Itoa(i)+".id")
		if err != nil {
			return nil, false, err
		}
		changed = true
	}
	return body, changed, nil
}
