package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/xchrisbradley/genagent/pkg/core"
	"github.com/xchrisbradley/genagent/plugins/ai"
	"github.com/xchrisbradley/genagent/plugins/ai/llm"
)

func main() {
	fmt.Println("\n=== GenAgent AI Chat Demo ===")

	// Get API keys from environment
	openaiKey := os.Getenv("OPENAI_API_KEY")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	if openaiKey == "" && geminiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY or GEMINI_API_KEY environment variable is required")
		return
	}

	// Create a new world
	world := core.NewWorld()

	// Create available provider
	var provider llm.Provider
	var err error

	switch {
	case openaiKey != "":
		provider = llm.NewOpenAIProvider(openaiKey, "")
	case geminiKey != "":
		provider, err = llm.NewGeminiProvider(context.Background(), geminiKey)
		if err != nil {
			fmt.Printf("Error creating Gemini provider: %v\n", err)
			return
		}
		defer provider.(*llm.GeminiProvider).Close()
	}

	// Create and validate the AI plugin
	aiPlugin := ai.NewPlugin(provider)

	if err := core.ValidatePlugin(aiPlugin); err != nil {
		fmt.Printf("Error validating plugin: %v\n", err)
		return
	}

	// Register plugin components first
	for _, component := range aiPlugin.Components() {
		world.RegisterComponent(reflect.TypeOf(component))
	}

	// Register plugin systems
	for _, system := range aiPlugin.Systems() {
		world.RegisterSystem(system)
	}

	// Create an agent
	agent := world.CreateEntity()

	// Initialize the plugin
	if err := aiPlugin.Initialize(world, agent); err != nil {
		fmt.Printf("Error initializing plugin: %v\n", err)
		return
	}

	fmt.Printf("\nUsing %s provider. Type 'exit' to quit.\n", provider.Name())
	fmt.Println("The AI will remember your conversation history.")
	fmt.Print("\nYou: ")

	// Get AI component
	aiType := reflect.TypeOf(&ai.Component{})
	aiComponent := world.GetComponent(agent, aiType).(*ai.Component)

	// Add initial system message
	aiComponent.Messages = append(aiComponent.Messages, llm.Message{
		Role:    "system",
		Content: "You are a helpful AI assistant. Be concise in your responses.",
	})

	// Start chat loop
	scanner := bufio.NewScanner(os.Stdin)
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	// Channel for user input
	inputChan := make(chan string)
	go func() {
		for scanner.Scan() {
			inputChan <- scanner.Text()
		}
	}()

	for {
		select {
		case input := <-inputChan:
			if strings.ToLower(input) == "exit" {
				fmt.Println("\nGoodbye!")
				return
			}

			aiComponent.LastMessage = input
			aiComponent.LastUpdate = time.Now()

		case <-ticker.C:
			world.Update(1.0 / 60.0)

			if aiComponent.LastResponse != "" {
				fmt.Print("\nYou: ")
				aiComponent.LastResponse = ""
			}
		}
	}
}
