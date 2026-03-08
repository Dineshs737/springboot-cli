package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents the SpringCLI MCP server.
type Server struct {
	mcpServer *server.MCPServer
}

// NewServer initializes the MCP server and registers the available tools.
func NewServer() *Server {
	// Create the MCP server definition
	s := server.NewMCPServer(
		"springboot-cli",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	mcpServer := &Server{
		mcpServer: s,
	}

	mcpServer.registerTools()
	return mcpServer
}

// Start begins listening for JSON-RPC messages on Stdio.
func (s *Server) Start(ctx context.Context) error {
	// stdio server blocks until Stdin is closed
	return server.ServeStdio(s.mcpServer)
}

// registerTools defines the capabilities the AI can use.
func (s *Server) registerTools() {
	// Tool 1: create_spring_project
	createTool := mcp.NewTool("create_spring_project",
		mcp.WithDescription("Create a new Spring Boot project. This runs 'springcli new' non-interactively."),
		mcp.WithString("name", mcp.Required(), mcp.Description("Name of the project folder to create")),
		mcp.WithString("group", mcp.DefaultString("com.example"), mcp.Description("Maven Group ID")),
		mcp.WithString("artifact", mcp.Description("Maven Artifact ID (defaults to name)")),
		mcp.WithString("java_version", mcp.DefaultString("21"), mcp.Description("Java version (e.g. 17, 21)")),
		mcp.WithString("language", mcp.DefaultString("java"), mcp.Description("Language: java, kotlin, or groovy")),
		mcp.WithString("build", mcp.DefaultString("maven"), mcp.Description("Build tool: maven, gradle, or gradle-kotlin")),
		mcp.WithString("dependencies", mcp.Description("Comma-separated list of Spring Boot starter dependencies (e.g. web,data-jpa,lombok)")),
	)
	s.mcpServer.AddTool(createTool, s.handleCreateProject)

	// Tool 2: run_spring_doctor
	doctorTool := mcp.NewTool("run_spring_doctor",
		mcp.WithDescription("Scans a Spring Boot project for common misconfigurations and Java version mismatches."),
		mcp.WithString("dir", mcp.Required(), mcp.DefaultString("."), mcp.Description("The target repository directory to scan.")),
	)
	s.mcpServer.AddTool(doctorTool, s.handleRunDoctor)

	// Tool 3: add_dependency
	addDepTool := mcp.NewTool("add_spring_dependency",
		mcp.WithDescription("Adds a Spring Boot dependency (starter) to the project's build file (pom.xml or build.gradle)."),
		mcp.WithString("dir", mcp.Required(), mcp.DefaultString("."), mcp.Description("The target repository directory.")),
		mcp.WithString("dependency", mcp.Required(), mcp.Description("The ID of the dependency (e.g. web, security, mysql@9.8.1).")),
	)
	s.mcpServer.AddTool(addDepTool, s.handleAddDependency)

	// Tool 4: generate_code
	generateTool := mcp.NewTool("generate_spring_component",
		mcp.WithDescription("Generates NestJS-style boilerplate code: modules, controllers, services, repositories."),
		mcp.WithString("dir", mcp.Required(), mcp.DefaultString("."), mcp.Description("The target repository directory.")),
		mcp.WithString("type", mcp.Required(), mcp.Description("The type of component to generate: module, controller, service, repository")),
		mcp.WithString("name", mcp.Required(), mcp.Description("The name of the component (can include subdirectory paths like admin/user)")),
	)
	s.mcpServer.AddTool(generateTool, s.handleGenerateComponent)

	// Tool 5: audit_dependencies
	auditTool := mcp.NewTool("audit_spring_dependencies",
		mcp.WithDescription("Audits explicit dependencies in the project via the OSV API to find vulnerabilities."),
		mcp.WithString("dir", mcp.Required(), mcp.DefaultString("."), mcp.Description("The target repository directory.")),
		mcp.WithBoolean("update", mcp.Description("If true, automatically updates vulnerable dependencies to their fixed versions in the build file.")),
	)
	s.mcpServer.AddTool(auditTool, s.handleAuditDependencies)
}

// --------------------------------------------------------------------------
// Tool Handlers
// --------------------------------------------------------------------------

// executeHelper runs the actual CLI commands by executing the binary itself,
// ensuring we capture the stdout correctly without breaking the JSON-RPC pipe.
// MCP Server runs in the same binary, but using os/exec prevents shared state issues
// and cleanly separates the MCP protocol from the CLI's rich UI print statements.
func executeCLI(args ...string) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not find executable: %w", err)
	}

	// We append an internal flag (if necessary in the future) to disable colors.
	// We'll rely on NO_COLOR env var which many CLI libraries respect.
	cmd := exec.Command(execPath, args...)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")

	out, err := cmd.CombinedOutput()
	outputStr := string(out)

	if err != nil {
		return fmt.Sprintf("Command failed: %s\nError: %v\nOutput: %s", args, err, outputStr), fmt.Errorf("cli execution error")
	}
	return outputStr, nil
}

func (s *Server) handleCreateProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := request.Params.Arguments.(map[string]interface{})
	name, _ := argsMap["name"].(string)
	group, _ := argsMap["group"].(string)
	artifact, _ := argsMap["artifact"].(string)
	javaVer, _ := argsMap["java_version"].(string)
	lang, _ := argsMap["language"].(string)
	build, _ := argsMap["build"].(string)
	deps, _ := argsMap["dependencies"].(string)

	args := []string{"new", name, "--no-interactive"}
	if group != "" {
		args = append(args, "--group", group)
	}
	if artifact != "" {
		args = append(args, "--artifact", artifact)
	}
	if javaVer != "" {
		args = append(args, "--java", javaVer)
	}
	if lang != "" {
		args = append(args, "--lang", lang)
	}
	if build != "" {
		args = append(args, "--build", build)
	}
	if deps != "" {
		args = append(args, "--deps", deps)
	}

	output, _ := executeCLI(args...)
	return mcp.NewToolResultText(fmt.Sprintf("```\n%s\n```", output)), nil
}

func (s *Server) handleRunDoctor(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := request.Params.Arguments.(map[string]interface{})
	dir, _ := argsMap["dir"].(string)
	output, _ := executeCLI("doctor", "--dir", dir)
	return mcp.NewToolResultText(fmt.Sprintf("```\n%s\n```", output)), nil
}

func (s *Server) handleAddDependency(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := request.Params.Arguments.(map[string]interface{})
	dir, _ := argsMap["dir"].(string)
	dep, _ := argsMap["dependency"].(string)
	output, _ := executeCLI("add", dep, "--dir", dir)
	return mcp.NewToolResultText(fmt.Sprintf("```\n%s\n```", output)), nil
}

func (s *Server) handleGenerateComponent(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := request.Params.Arguments.(map[string]interface{})
	dir, _ := argsMap["dir"].(string)
	compType, _ := argsMap["type"].(string)
	name, _ := argsMap["name"].(string)

	output, _ := executeCLI("generate", compType, name, "--dir", dir)
	return mcp.NewToolResultText(fmt.Sprintf("```\n%s\n```", output)), nil
}

func (s *Server) handleAuditDependencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	argsMap, _ := request.Params.Arguments.(map[string]interface{})
	dir, _ := argsMap["dir"].(string)
	update, ok := argsMap["update"].(bool)

	args := []string{"audit", "--dir", dir}
	if ok && update {
		args = append(args, "--update")
	}

	output, _ := executeCLI(args...)
	return mcp.NewToolResultText(fmt.Sprintf("```\n%s\n```", output)), nil
}
