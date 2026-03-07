package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Generator handles post-download project generation tasks.
type Generator struct{}

// NewGenerator creates a new Generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// InitGit runs git init, git add, and git commit in the given directory.
func (g *Generator) InitGit(dir string) error {
	commands := []struct {
		args []string
		msg  string
	}{
		{[]string{"git", "init"}, "initializing git repository"},
		{[]string{"git", "add", "."}, "staging files"},
		{[]string{"git", "commit", "-m", "chore: initial commit from SpringCLI"}, "creating initial commit"},
	}

	for _, c := range commands {
		cmd := exec.Command(c.args[0], c.args[1:]...)
		cmd.Dir = dir
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", c.msg, err)
		}
	}

	return nil
}

// GenerateEnvFile creates a .env.example file based on the selected dependencies.
func (g *Generator) GenerateEnvFile(dir string, deps []string) error {
	var sb strings.Builder

	sb.WriteString("# ═══════════════════════════════════════\n")
	sb.WriteString("# Application Configuration\n")
	sb.WriteString("# ═══════════════════════════════════════\n")
	sb.WriteString("SERVER_PORT=8080\n")
	sb.WriteString("SPRING_PROFILES_ACTIVE=dev\n")
	sb.WriteString("\n")

	depSet := make(map[string]bool)
	for _, d := range deps {
		depSet[strings.ToLower(d)] = true
	}

	// Database.
	if depSet["jpa"] || depSet["data-jpa"] || depSet["jdbc"] || depSet["postgresql"] {
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("# Database\n")
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("DB_HOST=localhost\n")
		sb.WriteString("DB_PORT=5432\n")
		sb.WriteString("DB_NAME=mydb\n")
		sb.WriteString("DB_USERNAME=postgres\n")
		sb.WriteString("DB_PASSWORD=password\n")
		sb.WriteString("SPRING_DATASOURCE_URL=jdbc:postgresql://${DB_HOST}:${DB_PORT}/${DB_NAME}\n")
		sb.WriteString("SPRING_DATASOURCE_USERNAME=${DB_USERNAME}\n")
		sb.WriteString("SPRING_DATASOURCE_PASSWORD=${DB_PASSWORD}\n")
		sb.WriteString("\n")
	}

	// MongoDB.
	if depSet["data-mongodb"] || depSet["mongodb"] {
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("# MongoDB\n")
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("MONGODB_HOST=localhost\n")
		sb.WriteString("MONGODB_PORT=27017\n")
		sb.WriteString("MONGODB_DATABASE=mydb\n")
		sb.WriteString("SPRING_DATA_MONGODB_URI=mongodb://${MONGODB_HOST}:${MONGODB_PORT}/${MONGODB_DATABASE}\n")
		sb.WriteString("\n")
	}

	// Redis.
	if depSet["data-redis"] || depSet["redis"] {
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("# Redis\n")
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("REDIS_HOST=localhost\n")
		sb.WriteString("REDIS_PORT=6379\n")
		sb.WriteString("SPRING_DATA_REDIS_HOST=${REDIS_HOST}\n")
		sb.WriteString("SPRING_DATA_REDIS_PORT=${REDIS_PORT}\n")
		sb.WriteString("\n")
	}

	// Kafka.
	if depSet["kafka"] {
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("# Kafka\n")
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("KAFKA_BOOTSTRAP_SERVERS=localhost:9092\n")
		sb.WriteString("SPRING_KAFKA_BOOTSTRAP_SERVERS=${KAFKA_BOOTSTRAP_SERVERS}\n")
		sb.WriteString("\n")
	}

	// RabbitMQ.
	if depSet["amqp"] || depSet["rabbitmq"] {
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("# RabbitMQ\n")
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("RABBITMQ_HOST=localhost\n")
		sb.WriteString("RABBITMQ_PORT=5672\n")
		sb.WriteString("SPRING_RABBITMQ_HOST=${RABBITMQ_HOST}\n")
		sb.WriteString("SPRING_RABBITMQ_PORT=${RABBITMQ_PORT}\n")
		sb.WriteString("\n")
	}

	// Security / JWT.
	if depSet["security"] || depSet["oauth2-resource-server"] {
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("# Security\n")
		sb.WriteString("# ═══════════════════════════════════════\n")
		sb.WriteString("JWT_SECRET=your-256-bit-secret-key-here\n")
		sb.WriteString("JWT_EXPIRATION=86400000\n")
		sb.WriteString("\n")
	}

	envPath := filepath.Join(dir, ".env.example")
	if err := os.WriteFile(envPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing .env.example: %w", err)
	}

	return nil
}

// GenerateDockerFiles creates Dockerfile, docker-compose.yml, and .dockerignore.
func (g *Generator) GenerateDockerFiles(dir string, buildTool string, javaVersion string, deps []string) error {
	if err := g.generateDockerfile(dir, buildTool, javaVersion); err != nil {
		return err
	}
	if err := g.generateDockerCompose(dir, deps); err != nil {
		return err
	}
	if err := g.generateDockerIgnore(dir); err != nil {
		return err
	}
	return nil
}

func (g *Generator) generateDockerfile(dir, buildTool, javaVersion string) error {
	var sb strings.Builder

	jdkTag := javaVersion + "-jdk-alpine"
	jreTag := javaVersion + "-jre-alpine"

	if buildTool == "maven" {
		sb.WriteString("# Build stage\n")
		sb.WriteString(fmt.Sprintf("FROM eclipse-temurin:%s AS builder\n", jdkTag))
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY .mvn/ .mvn/\n")
		sb.WriteString("COPY mvnw pom.xml ./\n")
		sb.WriteString("RUN chmod +x mvnw && ./mvnw dependency:go-offline -q\n")
		sb.WriteString("COPY src ./src\n")
		sb.WriteString("RUN ./mvnw clean package -DskipTests -q\n")
		sb.WriteString("\n")
		sb.WriteString("# Runtime stage\n")
		sb.WriteString(fmt.Sprintf("FROM eclipse-temurin:%s\n", jreTag))
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY --from=builder /app/target/*.jar app.jar\n")
		sb.WriteString("EXPOSE 8080\n")
		sb.WriteString("ENTRYPOINT [\"java\", \"-jar\", \"app.jar\"]\n")
	} else {
		sb.WriteString("# Build stage\n")
		sb.WriteString(fmt.Sprintf("FROM eclipse-temurin:%s AS builder\n", jdkTag))
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY gradle/ gradle/\n")
		sb.WriteString("COPY gradlew build.gradle* settings.gradle* ./\n")
		sb.WriteString("RUN chmod +x gradlew && ./gradlew dependencies --no-daemon -q\n")
		sb.WriteString("COPY src ./src\n")
		sb.WriteString("RUN ./gradlew clean bootJar --no-daemon -q\n")
		sb.WriteString("\n")
		sb.WriteString("# Runtime stage\n")
		sb.WriteString(fmt.Sprintf("FROM eclipse-temurin:%s\n", jreTag))
		sb.WriteString("WORKDIR /app\n")
		sb.WriteString("COPY --from=builder /app/build/libs/*.jar app.jar\n")
		sb.WriteString("EXPOSE 8080\n")
		sb.WriteString("ENTRYPOINT [\"java\", \"-jar\", \"app.jar\"]\n")
	}

	return os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte(sb.String()), 0644)
}

func (g *Generator) generateDockerCompose(dir string, deps []string) error {
	depSet := make(map[string]bool)
	for _, d := range deps {
		depSet[strings.ToLower(d)] = true
	}

	var sb strings.Builder
	sb.WriteString("version: '3.8'\n\n")
	sb.WriteString("services:\n")

	// App service.
	sb.WriteString("  app:\n")
	sb.WriteString("    build: .\n")
	sb.WriteString("    ports:\n")
	sb.WriteString("      - \"8080:8080\"\n")
	sb.WriteString("    env_file:\n")
	sb.WriteString("      - .env\n")

	var dependsOn []string

	// PostgreSQL.
	if depSet["jpa"] || depSet["data-jpa"] || depSet["jdbc"] || depSet["postgresql"] {
		dependsOn = append(dependsOn, "postgres")
	}
	// Redis.
	if depSet["data-redis"] || depSet["redis"] {
		dependsOn = append(dependsOn, "redis")
	}
	// MongoDB.
	if depSet["data-mongodb"] || depSet["mongodb"] {
		dependsOn = append(dependsOn, "mongodb")
	}
	// Kafka.
	if depSet["kafka"] {
		dependsOn = append(dependsOn, "kafka")
	}
	// RabbitMQ.
	if depSet["amqp"] || depSet["rabbitmq"] {
		dependsOn = append(dependsOn, "rabbitmq")
	}

	if len(dependsOn) > 0 {
		sb.WriteString("    depends_on:\n")
		for _, d := range dependsOn {
			sb.WriteString(fmt.Sprintf("      - %s\n", d))
		}
	}
	sb.WriteString("\n")

	// PostgreSQL service.
	if depSet["jpa"] || depSet["data-jpa"] || depSet["jdbc"] || depSet["postgresql"] {
		sb.WriteString("  postgres:\n")
		sb.WriteString("    image: postgres:16-alpine\n")
		sb.WriteString("    environment:\n")
		sb.WriteString("      POSTGRES_DB: mydb\n")
		sb.WriteString("      POSTGRES_USER: postgres\n")
		sb.WriteString("      POSTGRES_PASSWORD: password\n")
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"5432:5432\"\n")
		sb.WriteString("    volumes:\n")
		sb.WriteString("      - postgres_data:/var/lib/postgresql/data\n")
		sb.WriteString("\n")
	}

	// Redis service.
	if depSet["data-redis"] || depSet["redis"] {
		sb.WriteString("  redis:\n")
		sb.WriteString("    image: redis:7-alpine\n")
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"6379:6379\"\n")
		sb.WriteString("\n")
	}

	// MongoDB service.
	if depSet["data-mongodb"] || depSet["mongodb"] {
		sb.WriteString("  mongodb:\n")
		sb.WriteString("    image: mongo:7\n")
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"27017:27017\"\n")
		sb.WriteString("    volumes:\n")
		sb.WriteString("      - mongodb_data:/data/db\n")
		sb.WriteString("\n")
	}

	// Kafka + Zookeeper service.
	if depSet["kafka"] {
		sb.WriteString("  zookeeper:\n")
		sb.WriteString("    image: confluentinc/cp-zookeeper:7.6.0\n")
		sb.WriteString("    environment:\n")
		sb.WriteString("      ZOOKEEPER_CLIENT_PORT: 2181\n")
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"2181:2181\"\n")
		sb.WriteString("\n")
		sb.WriteString("  kafka:\n")
		sb.WriteString("    image: confluentinc/cp-kafka:7.6.0\n")
		sb.WriteString("    depends_on:\n")
		sb.WriteString("      - zookeeper\n")
		sb.WriteString("    environment:\n")
		sb.WriteString("      KAFKA_BROKER_ID: 1\n")
		sb.WriteString("      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181\n")
		sb.WriteString("      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092\n")
		sb.WriteString("      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1\n")
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"9092:9092\"\n")
		sb.WriteString("\n")
	}

	// RabbitMQ service.
	if depSet["amqp"] || depSet["rabbitmq"] {
		sb.WriteString("  rabbitmq:\n")
		sb.WriteString("    image: rabbitmq:3-management-alpine\n")
		sb.WriteString("    ports:\n")
		sb.WriteString("      - \"5672:5672\"\n")
		sb.WriteString("      - \"15672:15672\"\n")
		sb.WriteString("\n")
	}

	// Volumes.
	var volumes []string
	if depSet["jpa"] || depSet["data-jpa"] || depSet["jdbc"] || depSet["postgresql"] {
		volumes = append(volumes, "postgres_data")
	}
	if depSet["data-mongodb"] || depSet["mongodb"] {
		volumes = append(volumes, "mongodb_data")
	}

	if len(volumes) > 0 {
		sb.WriteString("volumes:\n")
		for _, v := range volumes {
			sb.WriteString(fmt.Sprintf("  %s:\n", v))
		}
	}

	return os.WriteFile(filepath.Join(dir, "docker-compose.yml"), []byte(sb.String()), 0644)
}

func (g *Generator) generateDockerIgnore(dir string) error {
	content := `.git
.gitignore
*.md
.env
.env.*
!.env.example
target/
build/
.gradle/
.mvn/
*.iml
.idea/
.vscode/
*.log
`
	return os.WriteFile(filepath.Join(dir, ".dockerignore"), []byte(content), 0644)
}
