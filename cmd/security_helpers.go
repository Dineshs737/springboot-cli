package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/springcli/springcli/internal/maven"
)

// detectKotlinProject checks if src/main/kotlin exists.
func detectKotlinProject(dir string) bool {
	kotlinDir := filepath.Join(dir, "src", "main", "kotlin")
	if info, err := os.Stat(kotlinDir); err == nil && info.IsDir() {
		return true
	}
	return false
}

// detectBasePackage derives the base package from the build file or source tree.
func detectBasePackage(dir, buildPath string, buildType interface{}) (string, error) {
	// Try Maven pom.xml.
	if strings.HasSuffix(buildPath, "pom.xml") {
		mp := maven.NewParser()
		pkg, err := mp.DetectBasePackage(buildPath)
		if err == nil && pkg != "" {
			return pkg, nil
		}
	}

	// Try finding the Application class in source dirs.
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

	return "com.example.demo", nil
}

// findApplicationClass walks the source tree to find *Application.java or *Application.kt.
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

// buildSecurityDir constructs the path for the security package directory.
func buildSecurityDir(dir, lang, basePackage string) string {
	packagePath := strings.ReplaceAll(basePackage, ".", string(os.PathSeparator))
	return filepath.Join(dir, "src", "main", lang, packagePath, "security")
}

// createDirIfNotExists creates a directory tree if it doesn't exist.
func createDirIfNotExists(dir string) error {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	return nil
}

// writeFile writes content to a file path.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// generateJavaSecurityConfig creates a Java SecurityConfig source file.
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

	sb.WriteString("\n@Configuration\n")
	sb.WriteString("@EnableWebSecurity\n")
	sb.WriteString("public class SecurityConfig {\n\n")

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

	sb.WriteString("    @Bean\n")
	sb.WriteString("    public PasswordEncoder passwordEncoder() {\n")
	sb.WriteString("        return new BCryptPasswordEncoder();\n")
	sb.WriteString("    }\n")

	sb.WriteString("}\n")
	return sb.String()
}

// generateKotlinSecurityConfig creates a Kotlin SecurityConfig source file.
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
