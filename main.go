package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/abcdlsj/kiwi/pkg/container"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// CORS 中间件配置
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"} // 允许前端域名
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	r.Use(cors.New(config))

	// API routes
	r.GET("/apps", listApps)
	r.GET("/apps/:name/template", getAppTemplate)
	r.POST("/apps/:name/deploy", deployApp)
	r.POST("/apps/:name/detact", detactAppConfig)

	r.Run(":8080")
}

func listApps(c *gin.Context) {
	// 从 service_template.yaml 文件加载服务模板
	template, err := container.LoadServicesTemplate("templates/service_template.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load service templates"})
		return
	}

	// 提取应用名称
	apps := make([]string, len(template.Services))
	for i, service := range template.Services {
		apps[i] = service.Name
	}

	c.JSON(http.StatusOK, gin.H{"apps": apps})
}

func getAppTemplate(c *gin.Context) {
	appName := c.Param("name")

	// 从 service_template.yaml 文件加载服务模板
	template, err := container.LoadServicesTemplate("templates/service_template.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load service templates"})
		return
	}

	// 查找指定的应用模板
	var appTemplate *container.ServiceConfig
	for _, service := range template.Services {
		if service.Name == appName {
			appTemplate = &service
			break
		}
	}

	if appTemplate == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "App template not found"})
		return
	}

	c.JSON(http.StatusOK, appTemplate)
}

func deployApp(c *gin.Context) {
	appName := c.Param("name")
	var deployConfig container.ServiceConfig
	if err := c.BindJSON(&deployConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deployOptions, err := deployConfig.ToServiceOptions()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create Docker deployer
	deployer, err := container.NewDockerDeployer()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writer := &container.DefaultOutputWriter{Writer: os.Stdout}

	// Deploy service
	err = deployer.DeployService(deployOptions, writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "App " + appName + " deployed successfully"})
}

func detactAppConfig(c *gin.Context) {
	appName := c.Param("name")
	var input struct {
		Message string `json:"message"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 加载默认模板
	template, err := container.LoadServicesTemplate("templates/service_template.yaml")
	if err != nil {
		log.Printf("Failed to load service template: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load service template"})
		return
	}

	var appConfig *container.ServiceConfig
	for _, service := range template.Services {
		if service.Name == appName {
			appConfig = &service
			break
		}
	}

	if appConfig == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "App template not found"})
		return
	}

	// 使用 OpenAI 检测配置变更
	updatedConfig, err := detectConfigChangesWithOpenAI(*appConfig, input.Message)
	if err != nil {
		log.Printf("Error detecting configuration: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error detecting configuration: %v", err)})
		return
	}

	response := gin.H{
		"message": "Detected configuration based on your input:",
		"options": updatedConfig,
	}

	c.JSON(http.StatusOK, response)
}

func detectConfigChangesWithOpenAI(config container.ServiceConfig, input string) (container.ServiceConfig, error) {
	openaiKey := os.Getenv("OPENAI_API_KEY")
	openaiEndpoint := os.Getenv("OPENAI_API_ENDPOINT")

	if openaiKey == "" || openaiEndpoint == "" {
		return config, fmt.Errorf("OpenAI API key or endpoint not set")
	}

	clientCfg := openai.DefaultConfig(openaiKey)
	clientCfg.BaseURL = openaiEndpoint
	client := openai.NewClientWithConfig(clientCfg)

	prompt := fmt.Sprintf(`
Given the following app configuration:
%s

And the user input:
"%s"

Please update the configuration according to the user input, if the user inputs a mount directory or port mapping, please remove the default placeholder port mapping and mount directory (e.g. /path/to/host or /path/to/container, or the default 8443:443 and 8080:80), and only return the updated JSON configuration without any explanation.
`, config, input)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that updates app configurations based on user input.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return config, fmt.Errorf("OpenAI API error: %v", err)
	}

	updatedConfigJSON := resp.Choices[0].Message.Content
	log.Printf("OpenAI response: %s", updatedConfigJSON)

	var updatedConfig container.ServiceConfig
	err = json.Unmarshal([]byte(updatedConfigJSON), &updatedConfig)
	if err != nil {
		return config, fmt.Errorf("error parsing OpenAI response: %v", err)
	}

	return updatedConfig, nil
}
