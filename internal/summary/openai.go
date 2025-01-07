package summary

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type OpenAISummarizer struct { // Структура для работы с chatGpt
	client  *openai.Client
	prompt  string
	enabled bool
	mu      sync.Mutex
}

func NewOpenAISummarizer(apiKey string, prompt string) *OpenAISummarizer {
	s := &OpenAISummarizer{
		client: openai.NewClient(apiKey), // NewClient создает новый OpenAI API клиент.
		prompt: prompt,
	}

	logrus.Errorf("openai summarizer enabled: %v", apiKey != "") // Логируем включение summarizer

	if apiKey != "" { // Если apiKey не пустой то Summarizer включен
		s.enabled = true
	}

	return s
}

func (s *OpenAISummarizer) Summarize(ctx context.Context, text string) (string, error) { // Метод Summarizе для получения саммари из chatGPT
	s.mu.Lock() // Блокируемся мьютексом так так библиотека не потокобезопасна
	defer s.mu.Unlock()

	if !s.enabled { // Если Summarizer выключен то возвращаем пустое Summary
		logrus.Errorf("Summarizer is disabled, can't generate summary")
		return "", nil
	}

	request := openai.ChatCompletionRequest{ // Создаем запрос к chatGPT
		Model: "gpt-3.5-turbo", // Модеть chatGPT
		Messages: []openai.ChatCompletionMessage{ // Слайс передоваемых сообщений
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("%s%s", text, s.prompt), // Передаем сам текст и просим сделать для него summary
			},
		},
		MaxTokens:   256,
		Temperature: 0.7,
		TopP:        1,
	}

	resp, err := s.client.CreateChatCompletion(ctx, request) // Вызов API для создания завершения сообщения чата.
	if err != nil {
		logrus.Errorf("Failed to to Create a completion for the chat message: %s", err)
		return "", err
	}

	rawSammary := strings.TrimSpace(resp.Choices[0].Message.Content) // Open ai вернет несколько вариантов, берем самый первый и избавляемся от лишних пробелов

	if strings.HasSuffix(rawSammary, ".") { // Проверяем сгененрировал ди chatGPT точку в конце статьи
		return rawSammary, nil
	}

	sentences := strings.Split(rawSammary, ".") // В ином случае разбиваем rawSammary на отдельные предложения

	return strings.Join(sentences[:len(sentences)-1], ".") + ".", nil // И джойним все предложения через точку и добавляем точку в конце
}
