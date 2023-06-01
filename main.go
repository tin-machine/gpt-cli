package main

import (
        "context"
        "flag"
        "fmt"
        "os"
        "io/ioutil"
        "time"
        "gopkg.in/yaml.v2"
        openai "github.com/sashabaranov/go-openai"
)

// PromptMapping is a struct to hold the yaml configuration
type PromptMapping struct {
        Prompts map[string]string `yaml:"prompts"`
}

func main() {
        // Define the command line option
        promptOption := flag.String("o", "", "The option for the prompt")
        addMessageFile := flag.String("a", "", "The option for the prompt")
        flag.Parse()
        fmt.Printf("addMessageFile: %v\n", *addMessageFile)

        // Read the yaml file
        yamlFile, err := ioutil.ReadFile("config.yaml")
        fmt.Println(string(yamlFile))

        if err != nil {
                fmt.Printf("yamlFile.Get err #%v ", err)
        }

        // Parse the yaml file
        var promptMapping PromptMapping
        err = yaml.Unmarshal(yamlFile, &promptMapping)
        if err != nil {
                fmt.Printf("Unmarshal: %v", err)
        }


        fmt.Printf("Prompt map: %v\n", promptMapping)

        // Get the prompt from the config based on the option
        prompt, ok := promptMapping.Prompts[*promptOption]
        if !ok {
                fmt.Printf("Prompt option %s is not defined in the config file", *promptOption)
                return
        }

        addMessage, err := ioutil.ReadFile(*addMessageFile)
        prompt = prompt + string(addMessage)

        client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
        resp, err := client.CreateChatCompletion(
                context.Background(),
                openai.ChatCompletionRequest{
                        Model: openai.GPT3Dot5Turbo, // Change this to GPT4 as per your requirement
                        Messages: []openai.ChatCompletionMessage{
                                {
                                        Role:    openai.ChatMessageRoleUser,
                                        Content: prompt,
                                },
                        },
                },
        )

        if err != nil {
                fmt.Printf("ChatCompletion error: %v\n", err)
                return
        }

        // Output the result
        fmt.Println(resp.Choices[0].Message.Content)

        // Save the conversation state to a directory named with Unix time
        dirName := fmt.Sprintf("%v", time.Now().Unix())
        err = os.Mkdir(dirName, 0755)
        if err != nil {
                fmt.Printf("Failed to create directory: %v\n", err)
                return
        }

        // Write the conversation to a file in the directory
        conversationFile := fmt.Sprintf("%s/conversation.txt", dirName)
        err = ioutil.WriteFile(conversationFile, []byte(resp.Choices[0].Message.Content), 0644)
        if err != nil {
                fmt.Printf("Failed to write conversation to file: %v\n", err)
                return
        }
}
