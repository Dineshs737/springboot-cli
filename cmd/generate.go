package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/springcli/springcli/internal/project"
)

var genDir string

var generateCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{"g"},
	Short:   "Generate Spring Boot code (NestJS style)",
	Long: `Generate boilerplate code for Spring Boot such as modules, controllers,
services, and repositories.

Examples:
  springcli generate module user
  springcli g controller product
  springcli g service order
  springcli g repository payment`,
}

var moduleCmd = &cobra.Command{
	Use:     "module <name>",
	Aliases: []string{"mo"},
	Short:   "Generate a new module (Controller, Service, Repository)",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate(args[0], true, true, true)
	},
}

var controllerCmd = &cobra.Command{
	Use:     "controller <name>",
	Aliases: []string{"co"},
	Short:   "Generate a Controller",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate(args[0], true, false, false)
	},
}

var serviceCmd = &cobra.Command{
	Use:     "service <name>",
	Aliases: []string{"s"},
	Short:   "Generate a Service",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate(args[0], false, true, false)
	},
}

var repositoryCmd = &cobra.Command{
	Use:     "repository <name>",
	Aliases: []string{"r"},
	Short:   "Generate a Repository",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate(args[0], false, false, true)
	},
}

func init() {
	generateCmd.PersistentFlags().StringVar(&genDir, "dir", ".", "Target project directory")

	generateCmd.AddCommand(moduleCmd)
	generateCmd.AddCommand(controllerCmd)
	generateCmd.AddCommand(serviceCmd)
	generateCmd.AddCommand(repositoryCmd)

	rootCmd.AddCommand(generateCmd)
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func parseName(input string) (string, string) {
	parts := strings.Split(input, "/")
	if len(parts) == 1 {
		return "", input // No subfolder
	}
	name := parts[len(parts)-1]
	subFolder := strings.Join(parts[:len(parts)-1], "/")
	return subFolder, name
}

func runGenerate(inputName string, genController, genService, genRepo bool) error {
	detector := project.NewDetector()
	buildPath, buildType, err := detector.GetBuildFilePath(genDir)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	isKotlin := detectKotlinProject(genDir)
	lang := "Java"
	if isKotlin {
		lang = "Kotlin"
	}

	rawBasePackage, err := detectBasePackage(genDir, buildPath, buildType)
	if err != nil {
		return fmt.Errorf("detecting base package: %w", err)
	}

	subFolder, name := parseName(strings.ToLower(inputName))
	classNamePrefix := capitalize(name)

	// Create the module package name
	var modulePackage string
	modulePathSegment := name
	if subFolder != "" {
		modulePackage = rawBasePackage + "." + strings.ReplaceAll(subFolder, "/", ".") + "." + name
		modulePathSegment = subFolder + "/" + name
	} else {
		modulePackage = rawBasePackage + "." + name
	}

	printer.Info("Generating %s code in %s", lang, modulePackage)

	langDir := "java"
	if isKotlin {
		langDir = "kotlin"
	}

	packagePath := strings.ReplaceAll(rawBasePackage, ".", string(os.PathSeparator))
	moduleDirPath := filepath.Join(genDir, "src", "main", langDir, packagePath, filepath.FromSlash(modulePathSegment))

	if err := createDirIfNotExists(moduleDirPath); err != nil {
		return fmt.Errorf("creating module directory: %w", err)
	}

	if genController {
		err = writeController(isKotlin, moduleDirPath, classNamePrefix, modulePackage, name)
		if err != nil {
			printer.Error("Failed to generate Controller: %v", err)
		} else {
			printer.Success("CREATE %sController", classNamePrefix)
		}
	}

	if genService {
		err = writeService(isKotlin, moduleDirPath, classNamePrefix, modulePackage)
		if err != nil {
			printer.Error("Failed to generate Service: %v", err)
		} else {
			printer.Success("CREATE %sService", classNamePrefix)
		}
	}

	if genRepo {
		err = writeRepository(isKotlin, moduleDirPath, classNamePrefix, modulePackage)
		if err != nil {
			printer.Error("Failed to generate Repository: %v", err)
		} else {
			printer.Success("CREATE %sRepository", classNamePrefix)
		}
	}

	return nil
}

func writeController(isKotlin bool, moduleDir, className, packageName, pathName string) error {
	var content string
	var err error
	if isKotlin {
		content = fmt.Sprintf(`package %s

import org.springframework.web.bind.annotation.*

@RestController
@RequestMapping("/%s")
class %sController(private val %sService: %sService) {


}
`, packageName, pathName, className, strings.ToLower(className), className)
		err = writeFile(filepath.Join(moduleDir, className+"Controller.kt"), content)
	} else {
		content = fmt.Sprintf(`package %s;

import org.springframework.web.bind.annotation.*;
import java.util.List;

@RestController
@RequestMapping("/%s")
public class %sController {

    private final %sService %sService;

    public %sController(%sService %sService) {
        this.%sService = %sService;
    }


}
`, packageName, pathName, className, className, strings.ToLower(className), className, className, strings.ToLower(className), strings.ToLower(className), strings.ToLower(className))
		err = writeFile(filepath.Join(moduleDir, className+"Controller.java"), content)
	}
	return err
}

func writeService(isKotlin bool, moduleDir, className, packageName string) error {
	var content string
	var err error
	if isKotlin {
		content = fmt.Sprintf(`package %s

import org.springframework.stereotype.Service

@Service
class %sService(private val %sRepository: %sRepository) {
}
`, packageName, className, strings.ToLower(className), className)
		err = writeFile(filepath.Join(moduleDir, className+"Service.kt"), content)
	} else {
		content = fmt.Sprintf(`package %s;

import org.springframework.stereotype.Service;

@Service
public class %sService {

    private final %sRepository %sRepository;

    public %sService(%sRepository %sRepository) {
        this.%sRepository = %sRepository;
    }
}
`, packageName, className, className, strings.ToLower(className), className, className, strings.ToLower(className), strings.ToLower(className), strings.ToLower(className))
		err = writeFile(filepath.Join(moduleDir, className+"Service.java"), content)
	}
	return err
}

func writeRepository(isKotlin bool, moduleDir, className, packageName string) error {
	var content string
	var err error
	if isKotlin {
		content = fmt.Sprintf(`package %s

import org.springframework.data.jpa.repository.JpaRepository
import org.springframework.stereotype.Repository

@Repository
interface %sRepository : JpaRepository<Any, Long> {
}
`, packageName, className)
		err = writeFile(filepath.Join(moduleDir, className+"Repository.kt"), content)
	} else {
		content = fmt.Sprintf(`package %s;

import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

@Repository
public interface %sRepository extends JpaRepository<Object, Long> {
}
`, packageName, className)
		err = writeFile(filepath.Join(moduleDir, className+"Repository.java"), content)
	}
	return err
}
