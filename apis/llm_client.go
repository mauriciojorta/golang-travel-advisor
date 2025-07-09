package apis

import (
	"context"
	"os"
	"strconv"
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
	temperatureStr := os.Getenv("LLM_TEMPERATURE")
	if temperatureStr == "" {
		log.Warn("LLM_TEMPERATURE environment variable is not set. Using default value of 0.6.")
		temperatureStr = "0.6"
	}

	temperature, err := strconv.ParseFloat(temperatureStr, 64)
	if err != nil {
		log.Warnf("Failed to parse LLM_TEMPERATURE: %v. Using default value of 0.6.", err)
		temperature = 0.6
	}
	log.Debugf("Using LLM temperature: %f", temperature)

	minLengthStr := os.Getenv("LLM_MIN_RESPONSE_LENGTH")
	if minLengthStr == "" {
		log.Warn("LLM_MIN_RESPONSE_LENGTH environment variable is not set. Using default value of 1500.")
		minLengthStr = "1500"
	}
	minLength, err := strconv.Atoi(minLengthStr)
	if err != nil {
		log.Warnf("Failed to parse LLM_MIN_RESPONSE_LENGTH: %v. Using default value of 1500.", err)
		minLength = 1500
	}
	log.Debugf("Using LLM minimum length: %d", minLength)

	maxLengthStr := os.Getenv("LLM_MAX_RESPONSE_LENGTH")
	if maxLengthStr == "" {
		log.Warn("LLM_MAX_RESPONSE_LENGTH environment variable is not set. Using default value of 3000.")
		maxLengthStr = "3000"
	}
	maxLength, err := strconv.Atoi(maxLengthStr)
	if err != nil {
		log.Warnf("Failed to parse LLM_MAX_RESPONSE_LENGTH: %v. Using default value of 3000.", err)
		maxLength = 3000
	}
	log.Debugf("Using LLM maximum length: %d", maxLength)

	response, err := llmClient.GenerateContent(ctx, messages, llms.WithTemperature(temperature), llms.WithMinLength(minLength), llms.WithMaxLength(maxLength))

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
