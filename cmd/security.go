package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/springcli/springcli/internal/gradle"
	"github.com/springcli/springcli/internal/maven"
	"github.com/springcli/springcli/internal/project"
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Configure Spring Security for the current project",
	Long: `Add Spring Security dependencies and optionally generate a
SecurityConfig class with JWT or OAuth2 support.`,
	RunE: runSecurity,
}

var (
	secJWT            bool
	secOAuth2         bool
	secGenerateConfig bool
	secDir            string
)

func init() {
	securityCmd.Flags().BoolVar(&secJWT, "jwt", false, "Add JWT (jjwt) dependencies")
	securityCmd.Flags().BoolVar(&secOAuth2, "oauth2", false, "Add OAuth2 resource server dependency")
	securityCmd.Flags().BoolVar(&secGenerateConfig, "generate-config", false, "Generate SecurityConfig source file")
	securityCmd.Flags().StringVar(&secDir, "dir", ".", "Target project directory")
}

func runSecurity(cmd *cobra.Command, args []string) error {
	detector := project.NewDetector()
	buildPath, buildType, err := detector.GetBuildFilePath(secDir)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	printer.Info("Detected %s project", buildType)

	// Always add spring-boot-starter-security.
	addSecDep(buildPath, buildType, "org.springframework.boot", "spring-boot-starter-security", "", "")

	if secJWT {
		printer.Info("Adding JWT (jjwt) dependencies...")
		addSecDep(buildPath, buildType, "io.jsonwebtoken", "jjwt-api", "0.12.5", "")
		addSecDep(buildPath, buildType, "io.jsonwebtoken", "jjwt-impl", "0.12.5", "runtime")
		addSecDep(buildPath, buildType, "io.jsonwebtoken", "jjwt-jackson", "0.12.5", "runtime")
	}

	if secOAuth2 {
		printer.Info("Adding OAuth2 resource server dependency...")
		addSecDep(buildPath, buildType, "org.springframework.boot", "spring-boot-starter-oauth2-resource-server", "", "")
	}

	if secGenerateConfig {
		if err := generateSecurityConfig(secDir, buildPath, buildType); err != nil {
			return fmt.Errorf("generating SecurityConfig: %w", err)
		}
	}

	return nil
}

func addSecDep(buildPath string, buildType project.BuildType, group, artifact, version, scope string) {
	var added bool
	var err error

	switch buildType {
	case project.BuildTypeMaven:
		mp := maven.NewParser()
		added, err = mp.AddDependency(buildPath, maven.MavenDependency{
			GroupID:    group,
			ArtifactID: artifact,
			Version:    version,
			Scope:      scope,
		})
	case project.BuildTypeGradleGroovy, project.BuildTypeGradleKotlin:
		gp := gradle.NewParser()
		added, err = gp.AddDependency(buildPath, group, artifact, version, scope)
	default:
		printer.Error("Unsupported build type: %s", buildType)
		return
	}

	if err != nil {
		printer.Error("Failed to add %s:%s: %v", group, artifact, err)
		return
	}

	if added {
		printer.Success("Added %s:%s", group, artifact)
	} else {
		printer.Warn("%s:%s already present", group, artifact)
	}
}

func generateSecurityConfig(dir, buildPath string, buildType project.BuildType) error {
	// Detect language (Java vs Kotlin).
	isKotlin := detectKotlinProject(dir)
	lang := "Java"
	if isKotlin {
		lang = "Kotlin"
	}
	printer.Info("Detected %s project, generating SecurityConfig...", lang)

	// Detect base package.
	basePackage, err := detectBasePackage(dir, buildPath, buildType)
	if err != nil {
		return fmt.Errorf("detecting base package: %w", err)
	}
	printer.Step("Base package: %s", basePackage)

	// Generate the config file.
	var configContent string
	var configPath string

	if isKotlin {
		configContent = generateKotlinSecurityConfig(basePackage)
		srcDir := filepath.Join(dir, "src", "main", "kotlin", strings.ReplaceAll(basePackage, ".", string(os.PathSeparator)), "security")
		if err := os.MkdirAll(srcDir, os.ModePerm); err != nil {
			return fmt.Errorf("creating security package directory: %w", err)
		}
		configPath = filepath.Join(srcDir, "SecurityConfig.kt")
	} else {
		configContent = generateJavaSecurityConfig(basePackage)
		srcDir := filepath.Join(dir, "src", "main", "java", strings.ReplaceAll(basePackage, ".", string(os.PathSeparator)), "security")
		if err := os.MkdirAll(srcDir, os.ModePerm); err != nil {
			return fmt.Errorf("creating security package directory: %w", err)
		}
		configPath = filepath.Join(srcDir, "SecurityConfig.java")
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("writing SecurityConfig: %w", err)
	}

	printer.Success("Generated SecurityConfig at %s", configPath)
	return nil
}

func detectKotlinProject(dir string) bool {
	kotlinDir := filepath.Join(dir, "src", "main", "kotlin")
	if info, err := os.Stat(kotlinDir); err == nil && info.IsDir() {
		return true
	}
	return false
}

func detectBasePackage(dir, buildPath string, buildType project.BuildType) (string, error) {
	if buildType == project.BuildTypeMaven {
		mp := maven.NewParser()
		return mp.DetectBasePackage(buildPath)
	}

	// For Gradle, try to find the Application class.
	srcDirs := []string{
		filepath.Join(dir, "src", "main", "java"),
		filepath.Join(dir, "src", "main", "kotlin"),
	}

	for _, srcDir := range srcDirs {
		pkg, err := findApplicationClass(srcDir)
		if err == nil && pkg != "" {
			return pkg, nil
		}
	}

	return "com.example.demo", nil // fallback default
}

func findApplicationClass(srcDir string) (string, error) {
	var foundPackage string
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		if strings.HasSuffix(name, "Application.java") || strings.HasSuffix(name, "Application.kt") {
			// Extract package from directory structure.
			rel, relErr := filepath.Rel(srcDir, filepath.Dir(path))
			if relErr != nil {
				return nil
			}
			foundPackage = strings.ReplaceAll(rel, string(os.PathSeparator), ".")
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return foundPackage, nil
}

func generateJavaSecurityConfig(basePackage string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("package %s.security;\n\n", basePackage))
	sb.WriteString("import org.springframework.context.annotation.Bean;\n")
	sb.WriteString("import org.springframework.context.annotation.Configuration;\n")
	sb.WriteString("import org.springframework.security.config.annotation.web.builders.HttpSecurity;\n")
	sb.WriteString("import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;\n")
	sb.WriteString("import org.springframework.security.config.http.SessionCreationPolicy;\n")
	sb.WriteString("import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;\n")
	sb.WriteString("import org.springframework.security.crypto.password.PasswordEncoder;\n")
	sb.WriteString("import org.springframework.security.web.SecurityFilterChain;\n")

	if secJWT {
		sb.WriteString("import org.springframework.security.oauth2.jwt.JwtDecoder;\n")
		sb.WriteString("import org.springframework.security.oauth2.jwt.NimbusJwtDecoder;\n")
	}
	if secOAuth2 {
		sb.WriteString("import org.springframework.security.oauth2.server.resource.authentication.JwtAuthenticationConverter;\n")
	}

	sb.WriteString("\n@Configuration\n")
	sb.WriteString("@EnableWebSecurity\n")
	sb.WriteString("public class SecurityConfig {\n\n")

	// SecurityFilterChain bean.
	sb.WriteString("    @Bean\n")
	sb.WriteString("    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {\n")
	sb.WriteString("        http\n")
	sb.WriteString("            .csrf(csrf -> csrf.disable())\n")
	sb.WriteString("            .sessionManagement(session -> session.sessionCreationPolicy(SessionCreationPolicy.STATELESS))\n")
	sb.WriteString("            .authorizeHttpRequests(auth -> auth\n")
	sb.WriteString("                .requestMatchers(\"/api/public/**\", \"/actuator/health\", \"/swagger-ui/**\", \"/v3/api-docs/**\").permitAll()\n")
	sb.WriteString("                .requestMatchers(\"/api/admin/**\").hasRole(\"ADMIN\")\n")
	sb.WriteString("                .anyRequest().authenticated()\n")
	sb.WriteString("            )")

	if secJWT || secOAuth2 {
		sb.WriteString("\n            .oauth2ResourceServer(oauth2 -> oauth2.jwt(jwt -> {}))")
	}

	sb.WriteString(";\n")
	sb.WriteString("        return http.build();\n")
	sb.WriteString("    }\n\n")

	// PasswordEncoder bean.
	sb.WriteString("    @Bean\n")
	sb.WriteString("    public PasswordEncoder passwordEncoder() {\n")
	sb.WriteString("        return new BCryptPasswordEncoder();\n")
	sb.WriteString("    }\n")

	sb.WriteString("}\n")
	return sb.String()
}

func generateKotlinSecurityConfig(basePackage string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("package %s.security\n\n", basePackage))
	sb.WriteString("import org.springframework.context.annotation.Bean\n")
	sb.WriteString("import org.springframework.context.annotation.Configuration\n")
	sb.WriteString("import org.springframework.security.config.annotation.web.builders.HttpSecurity\n")
	sb.WriteString("import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity\n")
	sb.WriteString("import org.springframework.security.config.http.SessionCreationPolicy\n")
	sb.WriteString("import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder\n")
	sb.WriteString("import org.springframework.security.crypto.password.PasswordEncoder\n")
	sb.WriteString("import org.springframework.security.web.SecurityFilterChain\n")

	sb.WriteString("\n@Configuration\n")
	sb.WriteString("@EnableWebSecurity\n")
	sb.WriteString("class SecurityConfig {\n\n")

	sb.WriteString("    @Bean\n")
	sb.WriteString("    fun securityFilterChain(http: HttpSecurity): SecurityFilterChain {\n")
	sb.WriteString("        http\n")
	sb.WriteString("            .csrf { it.disable() }\n")
	sb.WriteString("            .sessionManagement { it.sessionCreationPolicy(SessionCreationPolicy.STATELESS) }\n")
	sb.WriteString("            .authorizeHttpRequests { auth ->\n")
	sb.WriteString("                auth\n")
	sb.WriteString("                    .requestMatchers(\"/api/public/**\", \"/actuator/health\", \"/swagger-ui/**\", \"/v3/api-docs/**\").permitAll()\n")
	sb.WriteString("                    .requestMatchers(\"/api/admin/**\").hasRole(\"ADMIN\")\n")
	sb.WriteString("                    .anyRequest().authenticated()\n")
	sb.WriteString("            }")

	if secJWT || secOAuth2 {
		sb.WriteString("\n            .oauth2ResourceServer { it.jwt { } }")
	}

	sb.WriteString("\n        return http.build()\n")
	sb.WriteString("    }\n\n")

	sb.WriteString("    @Bean\n")
	sb.WriteString("    fun passwordEncoder(): PasswordEncoder = BCryptPasswordEncoder()\n")

	sb.WriteString("}\n")
	return sb.String()
}
