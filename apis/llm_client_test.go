package apis

import (
	"context"
	"os"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/llms"
)

// mockModel implements llms.Model for testing.
type mockModel struct {
	generateContentCalled bool
}

func (m *mockModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, opts ...llms.CallOption) (*llms.ContentResponse, error) {
	m.generateContentCalled = true
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content:    "Test completion",
				StopReason: "stop",
			},
		},
	}, nil
}

func (m *mockModel) Type() string { return "mock" }

// Call is required to satisfy the llms.Model interface.
func (m *mockModel) Call(ctx context.Context, prompt string, opts ...llms.CallOption) (string, error) {
	m.generateContentCalled = true
	return "Test completion", nil
}

func resetSingleton() {
	llmClient = nil
	llmOnce = sync.Once{}
}

func TestInitLlmClient_Defaults(t *testing.T) {
	resetSingleton()
	os.Setenv("OPENAI_API_KEY", "test_key")
	os.Unsetenv("LLM_VENDOR")
	os.Unsetenv("LLM_MODEL")
	defer os.Clearenv()

	err := InitLlmClient()
	assert.NoError(t, err)
}

func TestInitLlmClient_WithOpenAIVendor(t *testing.T) {
	resetSingleton()
	os.Setenv("OPENAI_API_KEY", "test_key")
	os.Setenv("LLM_VENDOR", "openai")
	os.Unsetenv("LLM_MODEL")
	defer os.Clearenv()

	err := InitLlmClient()
	assert.NoError(t, err)
}

func TestInitLlmClient_WithModelEnvAndDefaultVendor(t *testing.T) {
	resetSingleton()
	os.Setenv("OPENAI_API_KEY", "test_key")
	os.Unsetenv("LLM_VENDOR")
	os.Setenv("LLM_MODEL", "gpt-4")
	defer os.Clearenv()

	err := InitLlmClient()
	assert.NoError(t, err)
}

func TestInitLlmClient_DefaultVendorWithoutOpenAIKey(t *testing.T) {
	resetSingleton()
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("LLM_VENDOR")
	defer os.Clearenv()

	err := InitLlmClient()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing the OpenAI API key")
}

func TestInitLlmClient_OpenAIVendorWithoutOpenAIKey(t *testing.T) {
	resetSingleton()
	os.Unsetenv("OPENAI_API_KEY")
	os.Setenv("LLM_VENDOR", "openai")
	defer os.Clearenv()

	err := InitLlmClient()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing the OpenAI API key")
}

func TestCallLlm_UsesLlmClient(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test_key")
	defer os.Clearenv()

	// Replace llmClient with a mock
	resetSingleton()
	mock := &mockModel{}
	llmClient = mock

	// CallLlm prints to stdout and may call log.Fatal on error, so we only test that GenerateContent is called.
	// To avoid os.Exit from log.Fatal, we do not trigger error paths.
	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "You are a helpful expert of international travel."},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "What is the capital of France?"},
			},
		},
	}

	response, _ := CallLlm(messages)
	assert.True(t, mock.generateContentCalled)
	assert.NotEmpty(t, response)

}

func TestCallRealLlm_UsesLlmClient(t *testing.T) {
	t.Skip("Skipping real LLM call test, requires environment setup and API key")

	log.SetLevel(log.DebugLevel) // Set log level to avoid debug output during tests

	err := godotenv.Load("../.env")
	assert.NoError(t, err)

	err = InitLlmClient()
	assert.NoError(t, err)

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "You are a helpful expert of international travel."},
			},
		},
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "What is the capital of France?"},
			},
		},
	}
	response, err := CallLlm(messages)
	assert.NoError(t, err)
	assert.NotEmpty(t, response)
}

func TestCallLlm_DefaultParameters(t *testing.T) {
	resetSingleton()
	os.Setenv("OPENAI_API_KEY", "test_key")
	defer os.Clearenv()

	mock := &mockModel{}
	llmClient = mock

	os.Unsetenv("LLM_TEMPERATURE")
	os.Unsetenv("LLM_MIN_RESPONSE_LENGTH")
	os.Unsetenv("LLM_MAX_RESPONSE_LENGTH")

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Test default parameters."},
			},
		},
	}

	resp, err := CallLlm(messages)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, mock.generateContentCalled)
}

func TestCallLlm_CustomParameters(t *testing.T) {
	resetSingleton()
	os.Setenv("OPENAI_API_KEY", "test_key")
	defer os.Clearenv()

	mock := &mockModel{}
	llmClient = mock

	os.Setenv("LLM_TEMPERATURE", "0.2")
	os.Setenv("LLM_MIN_RESPONSE_LENGTH", "123")
	os.Setenv("LLM_MAX_RESPONSE_LENGTH", "456")

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Test custom parameters."},
			},
		},
	}

	resp, err := CallLlm(messages)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, mock.generateContentCalled)
}

func TestCallLlm_InvalidTemperature(t *testing.T) {
	resetSingleton()
	os.Setenv("OPENAI_API_KEY", "test_key")
	defer os.Clearenv()

	mock := &mockModel{}
	llmClient = mock

	os.Setenv("LLM_TEMPERATURE", "not-a-number")

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Test invalid temperature."},
			},
		},
	}

	resp, err := CallLlm(messages)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, mock.generateContentCalled)
}

func TestCallLlm_InvalidMinLength(t *testing.T) {
	resetSingleton()
	os.Setenv("OPENAI_API_KEY", "test_key")
	defer os.Clearenv()

	mock := &mockModel{}
	llmClient = mock

	os.Setenv("LLM_MIN_RESPONSE_LENGTH", "not-a-number")

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Test invalid min length."},
			},
		},
	}

	resp, err := CallLlm(messages)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, mock.generateContentCalled)
}

func TestCallLlm_InvalidMaxLength(t *testing.T) {
	resetSingleton()
	os.Setenv("OPENAI_API_KEY", "test_key")
	defer os.Clearenv()

	mock := &mockModel{}
	llmClient = mock

	os.Setenv("LLM_MAX_RESPONSE_LENGTH", "not-a-number")

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Test invalid max length."},
			},
		},
	}

	resp, err := CallLlm(messages)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, mock.generateContentCalled)
}
