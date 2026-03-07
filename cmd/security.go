package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/springcli/springcli/internal/gradle"
	"github.com/springcli/springcli/internal/maven"
	"github.com/springcli/springcli/internal/project"
	"github.com/springcli/springcli/internal/wizard"
)

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Configure Spring Security for the current project",
	Long: `Add Spring Security dependencies and optionally generate a
SecurityConfig class with JWT or OAuth2 support.
Runs as a mini-wizard when interactive, or uses flags for CI/CD.`,
	RunE: runSecurity,
}

var (
	secJWT            bool
	secOAuth2         bool
	secGenerateConfig bool
	secDir            string
	secNoInteractive  bool
)

func init() {
	securityCmd.Flags().BoolVar(&secJWT, "jwt", false, "Add JWT (jjwt) dependencies")
	securityCmd.Flags().BoolVar(&secOAuth2, "oauth2", false, "Add OAuth2 resource server dependency")
	securityCmd.Flags().BoolVar(&secGenerateConfig, "generate-config", false, "Generate SecurityConfig source file")
	securityCmd.Flags().StringVar(&secDir, "dir", ".", "Target project directory")
	securityCmd.Flags().BoolVar(&secNoInteractive, "no-interactive", false, "Skip interactive wizard")
}

func runSecurity(cmd *cobra.Command, args []string) error {
	detector := project.NewDetector()
	buildPath, buildType, err := detector.GetBuildFilePath(secDir)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	printer.Info("Detected %s project", buildType)

	// Interactive wizard or flags.
	if !secNoInteractive && wizard.IsInteractive() {
		secType, genConfig, wizErr := wizard.RunSecurityWizard()
		if wizErr != nil {
			return fmt.Errorf("security wizard: %w", wizErr)
		}

		switch secType {
		case "JWT (Stateless API)":
			secJWT = true
		case "OAuth2 Resource Server":
			secOAuth2 = true
		case "JWT + OAuth2":
			secJWT = true
			secOAuth2 = true
		}
		secGenerateConfig = genConfig
	}

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
	isKotlin := detectKotlinProject(dir)
	lang := "Java"
	if isKotlin {
		lang = "Kotlin"
	}
	printer.Info("Detected %s project, generating SecurityConfig...", lang)

	basePackage, err := detectBasePackage(dir, buildPath, buildType)
	if err != nil {
		return fmt.Errorf("detecting base package: %w", err)
	}
	printer.Step("Base package: %s", basePackage)

	var configContent string
	var configPath string

	if isKotlin {
		configContent = generateKotlinSecurityConfig(basePackage)
		srcDir := buildSecurityDir(dir, "kotlin", basePackage)
		if err := createDirIfNotExists(srcDir); err != nil {
			return err
		}
		configPath = srcDir + string('/') + "SecurityConfig.kt"
	} else {
		configContent = generateJavaSecurityConfig(basePackage)
		srcDir := buildSecurityDir(dir, "java", basePackage)
		if err := createDirIfNotExists(srcDir); err != nil {
			return err
		}
		configPath = srcDir + string('/') + "SecurityConfig.java"
	}

	if err := writeFile(configPath, configContent); err != nil {
		return fmt.Errorf("writing SecurityConfig: %w", err)
	}

	printer.Success("Generated SecurityConfig at %s", configPath)
	return nil
}
