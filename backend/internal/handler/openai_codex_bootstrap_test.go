package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const testCodexDelegationEnvelope = `<codex_delegation><source_thread_id>thread-1</source_thread_id><input>do the work</input></codex_delegation>`

func TestNormalizeCodexDelegationBootstrap(t *testing.T) {
	body := []byte(`{"model":"gpt-5","metadata":{"integer":9007199254740993},"input":[{"type":"function_call_output","namespace":"codex_tui","name":"send_message_to_thread","output":` + testJSON(t, testCodexDelegationEnvelope) + `}]}`)

	got, changed := normalizeCodexDelegationBootstrap(body)
	require.True(t, changed)
	require.Equal(t, "message", gjson.GetBytes(got, "input.0.type").String())
	require.Equal(t, "user", gjson.GetBytes(got, "input.0.role").String())
	require.Equal(t, testCodexDelegationEnvelope, gjson.GetBytes(got, "input.0.content.0.text").String())
	require.Equal(t, "9007199254740993", gjson.GetBytes(got, "metadata.integer").Raw)
}

func TestNormalizeCodexDelegationBootstrapRejectsAmbiguousOrInvalidInput(t *testing.T) {
	tests := []string{
		`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"create_thread","call_id":"call-1","output":` + testJSON(t, testCodexDelegationEnvelope) + `}]}`,
		`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"create_thread","output":` + testJSON(t, testCodexDelegationEnvelope) + `},{"type":"computer_call"}]}`,
		`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"create_thread","output":"not an envelope"}]}`,
		`{"model":"gpt-5","input":[{"type":"message","type":"function_call_output","namespace":"codex_app","name":"create_thread","output":` + testJSON(t, testCodexDelegationEnvelope) + `}]}`,
	}
	for _, input := range tests {
		got, changed := normalizeCodexDelegationBootstrap([]byte(input))
		require.False(t, changed)
		require.Equal(t, input, string(got))
	}
}

func TestNormalizeCodexDelegationBootstrapRequiresCompleteEnvelope(t *testing.T) {
	for _, envelope := range []string{
		`<codex_delegation><source_thread_id>thread-1</source_thread_id></codex_delegation>`,
		`<codex_delegation><source_thread_id></source_thread_id><input>work</input></codex_delegation>`,
		`prefix<codex_delegation><source_thread_id>thread-1</source_thread_id><input>work</input></codex_delegation>`,
		`<codex_delegation><source_thread_id>thread-1</source_thread_id><input>work</input><extra>x</extra></codex_delegation>`,
		`<codex_delegation><x:source_thread_id>thread-1</x:source_thread_id><input>work</input></codex_delegation>`,
	} {
		body := []byte(`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"create_thread","output":` + testJSON(t, envelope) + `}]}`)
		got, changed := normalizeCodexDelegationBootstrap(body)
		require.Falsef(t, changed, "envelope %q", envelope)
		require.Equal(t, body, got)
	}
}

func TestNormalizeCodexDelegationBootstrapPreservesOrderAndIsIdempotent(t *testing.T) {
	body := []byte(`{"model":"gpt-5","input":[{"type":"message","role":"user","content":"before"},{"type":"function_call_output","namespace":"codex_app","name":"create_thread","output":` + testJSON(t, testCodexDelegationEnvelope) + `},{"type":"message","role":"user","content":"after"}]}`)

	got, changed := normalizeCodexDelegationBootstrap(body)
	require.True(t, changed)
	require.Equal(t, "before", gjson.GetBytes(got, "input.0.content").String())
	require.Equal(t, testCodexDelegationEnvelope, gjson.GetBytes(got, "input.1.content.0.text").String())
	require.Equal(t, "after", gjson.GetBytes(got, "input.2.content").String())

	again, changedAgain := normalizeCodexDelegationBootstrap(got)
	require.False(t, changedAgain)
	require.Equal(t, got, again)
}

func TestNormalizeCodexAutomationBootstrap(t *testing.T) {
	output := "Automation: Scheduled project review\n" +
		"Automation ID: wiki-maintenance\n" +
		"Automation memory: $CODEX_HOME/automations/wiki-maintenance/memory.md\n" +
		"Last run: 2026-09-01T02:06:34.536Z (1788228394536)\n\n" +
		"Review the project and report any important changes."
	body := []byte(`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"automation_update","output":` + testJSON(t, output) + `}]}`)

	got, changed := normalizeCodexAutomationBootstrap(body)
	require.True(t, changed)
	require.Equal(t, "message", gjson.GetBytes(got, "input.0.type").String())
	require.Equal(t, output, gjson.GetBytes(got, "input.0.content.0.text").String())
}

func TestNormalizeCodexAutomationBootstrapRejectsUnsafeInput(t *testing.T) {
	validOutput := "Automation: Scheduled project review\n" +
		"Automation ID: wiki\n" +
		"Automation memory: $CODEX_HOME/automations/wiki/memory.md\n" +
		"Last run: never\n\n" +
		"Review the project."
	tests := []string{
		`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"automation_update","call_id":"call-1","output":` + testJSON(t, validOutput) + `}]}`,
		`{"model":"gpt-5","previous_response_id":"resp-1","input":[{"type":"function_call_output","namespace":"codex_app","name":"automation_update","output":` + testJSON(t, validOutput) + `}]}`,
		`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"automation_update","output":` + testJSON(t, validOutput) + `},{"type":"computer_call_output","output":"done"}]}`,
		`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"automation_update","output":"Automation: x\nAutomation ID: ../wiki\nAutomation memory: $CODEX_HOME/automations/../wiki/memory.md\nLast run: never\n\nprompt"}]}`,
	}
	for _, input := range tests {
		got, changed := normalizeCodexAutomationBootstrap([]byte(input))
		require.False(t, changed)
		require.Equal(t, input, string(got))
	}
}

func TestNormalizeCodexAutomationBootstrapSupportsCRLFAndRejectsInvalidLastRun(t *testing.T) {
	validOutput := "Automation: Scheduled project review\n" +
		"Automation ID: wiki\n" +
		"Automation memory: $CODEX_HOME/automations/wiki/memory.md\n" +
		"Last run: never\n\n" +
		"Review the project."
	crlfOutput := strings.ReplaceAll(validOutput, "\n", "\r\n")
	body := []byte(`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"automation_update","output":` + testJSON(t, crlfOutput) + `}]}`)
	got, changed := normalizeCodexAutomationBootstrap(body)
	require.True(t, changed)
	require.Equal(t, crlfOutput, gjson.GetBytes(got, "input.0.content.0.text").String())

	invalidLastRun := strings.Replace(validOutput, "Last run: never", "Last run: yesterday", 1)
	invalidBody := []byte(`{"model":"gpt-5","input":[{"type":"function_call_output","namespace":"codex_app","name":"automation_update","output":` + testJSON(t, invalidLastRun) + `}]}`)
	got, changed = normalizeCodexAutomationBootstrap(invalidBody)
	require.False(t, changed)
	require.Equal(t, invalidBody, got)
}

func testJSON(t *testing.T, value string) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return string(encoded)
}
