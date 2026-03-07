<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/Spring%20Boot-CLI-6DB33F?style=for-the-badge&logo=springboot&logoColor=white" alt="Spring Boot" />
  <img src="https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge" alt="License" />
  <img src="https://img.shields.io/badge/Platform-Linux%20|%20macOS%20|%20Windows-blue?style=for-the-badge" alt="Platform" />
</p>

<h1 align="center">🌱 SpringCLI</h1>

<p align="center">
  <strong>The ultimate terminal experience for Spring Boot developers.</strong><br/>
  Create projects, manage dependencies, and configure security—all without leaving your terminal.
</p>

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-how-to-use-springcli">How to Use</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-advanced-usage-cicd">CI/CD & Scripts</a>
</p>

---

```
 ██████  ██████  ██████  ██ ███    ██  ██████   ██████ ██      ██
 ██      ██   ██ ██   ██ ██ ████   ██ ██       ██      ██      ██
 ╚█████  ██████  ██████  ██ ██ ██  ██ ██   ███ ██      ██      ██
      ██ ██      ██   ██ ██ ██  ██ ██ ██    ██ ██      ██      ██
 ██████  ██      ██   ██ ██ ██   ████  ██████   ██████ ███████ ██

 The Spring Boot CLI — like create-next-app, but for Java/Kotlin
```

## ⚡ Quick Start

With `springcli` installed (see [Installation](#-installation)), creating a new Spring Boot application is exactly one command:

```bash
springcli new my-awesome-api
```

It will launch a beautiful interactive wizard guiding you through Java vs Kotlin, Maven vs Gradle, and live-fetching the latest dependency catalog from `start.spring.io`.

Once built, you can run right away:
```bash
cd my-awesome-api
./mvnw spring-boot:run
```

---

## 📘 How to Use SpringCLI

SpringCLI does three major things perfectly: Scaffold projects, manage dependencies (like `npm`), and wire up Spring Security instantly. Here is exactly how to dominate your Spring Boot development with them.

### 1. Creating Projects (The Interactive Wizard)
Stop going to web browsers to download ZIP files.

```bash
springcli new my-service
```
This triggers the **Interactive Wizard**. It asks you everything Spring Initializr would ask: Group IDs, Packaging (JAR/WAR), Languages, and Dependencies.

**The Post-Generation Magic:**
Unlike the website, SpringCLI goes three steps further locally after the download finishes:
1. **Git Initialization:** Runs `git init` and commits your initial codebase.
2. **Containerization Engine:** Automatically scaffolds a multi-stage `Dockerfile` and creates a `.dockerignore`.
3. **Environment Generation:** It spots what dependencies you chose and generates a `.env.example` file. (e.g., if you picked `jpa` or `postgresql`, it writes out `DB_HOST`, `DB_PORT`, `SPRING_DATASOURCE_URL` for you). It will even build a `docker-compose.yml` defining those services for local `.env` usage!

*(If you ever want to skip the wizard and fully script the creation in bash, check the [CI/CD instructions below](#-advanced-usage-cicd)).*

### 2. Managing Dependencies (The `npm` of Java)
Have you ever forgotten exactly what `groupId:artifactId` was required for Spring Kafka? Stop searching Maven Central.

Use `springcli add` to resolve artifacts naturally:
```bash
# Need a web server?
springcli add web

# Need multiple dependencies at once?
springcli add jpa postgresql redis kafka
```
SpringCLI automatically detects if you are using `pom.xml`, `build.gradle` (Groovy), or `build.gradle.kts` (Kotlin), calls out to Spring Initializr's metadata library once, resolves the proper group and artifacts, and seamlessly patches your build files without messing up your indentation!

If you make a typo (`springcli add kafa`), it uses fuzzy-searching to suggest the right one ("Did you mean: kafka?").

**Removing or Viewing Installed Dependencies:**
```bash
# Remove exactly as you added 
springcli remove web
springcli rm jpa redis   # shorter alias works too

# Browse the entire Spring Boot catalog interactively
springcli list

# View only what is currently installed in your project
springcli list --installed
```

### 3. Setting Up Security Fast
Wiring up Spring Security always involves clicking around for tutorials on how to write `SecurityFilterChain`. One command does it all now:

```bash
springcli security
```
It asks you if you want Basic HTTP, JWT (Stateless), or OAuth2, downloads the correct dependencies naturally, **auto-detects if your project is Java or Kotlin**, discovers your exact Base Package name by scanning your `src/main` directory, and **generates a complete `SecurityConfig` source file** with CSRF protection, password encoders, and public whitelist routes out of the box.

---

## 📦 Installation

Grab the executable native binary for your architecture. SpringCLI has zero external dependencies (does not require JVM or Docker to run). 

### Linux / macOS (Automated Script)
The easiest way to install on UNIX systems is using the installation script. It will automatically detect your OS and architecture, download the latest release, and move it to `/usr/local/bin`.

```bash
curl -fsSL https://raw.githubusercontent.com/Dineshs737/springboot-cli/main/install.sh | bash
```

<details>
<summary>Manual Installation (Linux/macOS)</summary>

```bash
# macOS (Apple Silicon M1/M2/M3)
curl -Lo springcli https://github.com/Dineshs737/springboot-cli/releases/latest/download/springcli-darwin-arm64
chmod +x springcli
sudo mv springcli /usr/local/bin/

# macOS (Intel)
curl -Lo springcli https://github.com/Dineshs737/springboot-cli/releases/latest/download/springcli-darwin-amd64
chmod +x springcli
sudo mv springcli /usr/local/bin/

# Linux (amd64)
curl -Lo springcli https://github.com/Dineshs737/springboot-cli/releases/latest/download/springcli-linux-amd64
chmod +x springcli
sudo mv springcli /usr/local/bin/
```
</details>

### Windows (PowerShell)
```powershell
Invoke-WebRequest -Uri "https://github.com/Dineshs737/springboot-cli/releases/latest/download/springcli-windows-amd64.exe" -OutFile "springcli.exe"
# Move to a PATH folder, e.g., WindowsApps
Move-Item springcli.exe "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps\"
```

### Go Developers (Build from source)
```bash
go install github.com/springcli/springcli@latest
```

---

## 🤖 Advanced Usage (CI/CD)

SpringCLI fully detects if stdin is a pseudo-TTY terminal. If it detects it's trapped in a bash script or GitHub Actions, it automatically behaves completely non-interactively. However, you explicitly trigger this utilizing the `--no-interactive` flag.

Generate a full production microservice programmatically:
```bash
springcli new order-service \
  --no-interactive \
  --group com.acme.corp \
  --artifact order-api \
  --language kotlin \
  --type gradle-kotlin \
  --java 21 \
  --boot 3.3.0 \
  --deps web,jpa,security,actuator \
  --git \
  --docker \
  --env
```

Script security additions rapidly:
```bash
springcli security --jwt --generate-config --no-interactive
```

---

## 🏗 Architecture & Under the Hood

> For contributors and the ultra-curious

SpringCLI is written purely in Go 1.22+ to provide lightning-fast, zero-dependency executions locally.

*   **CLI UX Framework:** Powered by `spf13/cobra` (the same engine behind Kubernetes `kubectl` and `gh` cli) mixed with `AlecAivazis/survey/v2` for beautiful terminal wizards.
*   **Networking:** Pure Go stdlib `net/http` to communicate directly with JSON outputs of the Spring Initializr Metadata API endpoint (`/metadata/client`). State and Boot versions are fetched completely in real-time, meaning SpringCLI never gets "out of date".
*   **File Modifications:** It manipulates Maven XML trees natively utilizing `beevik/etree` to ensure that `<dependency>` nodes injected stay perfectly indented and do not destroy your custom Maven profiles. (Gradle builds utilize robust regex and line-level text modifications matching either `.groovy` or `.kts` syntaxes). 

### Continuous Integration
Continuous deployments compile cross-platform artifacts statically across 6 targets (Linux, Mac, Windows x AMD64, ARM64) with `CGO_ENABLED=0` to assure the produced `springcli` commands can execute entirely air-gapped from C libs. 

### Tests
We pride ourselves on zero-flakiness tests:
```bash
go test ./... -v -race          # Run all tests locally
```

## 📄 License & Contributing

Built with ❤️ by [Dinesh](https://github.com/Dineshs737).

Feature requests? Bug spotted? PRs are incredibly welcomed!
Licensed under the MIT License — see [LICENSE](LICENSE).
