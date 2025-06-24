package apis

import (
	"context"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

var (
	llmClient llms.Model
	llmOnce   sync.Once
)

// initLlmClient initializes llmClient as a singleton.
func InitLlmClient() error {
	var err error
	llmOnce.Do(func() {
		llm_vendor := os.Getenv("LLM_VENDOR")

		if llm_vendor == "" {
			log.Warn("LLM_VENDOR environment variable is not set. Using OpenAI by default.")
			llm_model := os.Getenv("LLM_MODEL")
			if llm_model == "" {
				log.Warn("LLM_MODEL environment variable is not set. Using gpt-3.5-turbo by default.")
				llm_model = "gpt-3.5-turbo"
			}

			llmClient, err = openai.New(openai.WithModel(llm_model))
			return
		}

		if llm_vendor == "openai" {
			log.Info("LLM_VENDOR is OpenAI")
			llm_model := os.Getenv("LLM_MODEL")

			if llm_model == "" {
				log.Warn("LLM_MODEL environment variable is not set. Using gpt-3.5-turbo by default.")
				llm_model = "gpt-3.5-turbo"
			} else {
				log.Infof("Using OpenAI model: %s", llm_model)
			}

			llmClient, err = openai.New(openai.WithModel(llm_model))
			return
		}
	})

	return err
}

var CallLlm = func(messages []llms.MessageContent) (*string, error) {

	ctx := context.Background()
	response, err := llmClient.GenerateContent(ctx, messages, llms.WithTemperature(0.8), llms.WithMinLength(1500), llms.WithMaxLength(2000))

	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Info("LLM request completed successfully")
	if len(response.Choices) < 1 {
		log.Warn("LLM response contains no choices")
	} else {
		log.Debugf("LLM response: %s", response.Choices[0].Content)
		return &response.Choices[0].Content, nil
	}

	return nil, nil

}
