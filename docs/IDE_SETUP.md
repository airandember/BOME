# IDE/Editor Setup for BOME Project

## Recommended Editor: Visual Studio Code (VS Code)

### Extensions
- **Svelte for VS Code** (svelte.svelte-vscode)
- **ESLint** (dbaeumer.vscode-eslint)
- **Prettier - Code formatter** (esbenp.prettier-vscode)
- **Tailwind CSS IntelliSense** (bradlc.vscode-tailwindcss)
- **Go** (golang.go)
- **Docker** (ms-azuretools.vscode-docker)
- **GitLens** (eamodio.gitlens)
- **REST Client** (humao.rest-client)
- **EditorConfig for VS Code** (editorconfig.editorconfig)

### Recommended Settings (add to `.vscode/settings.json`)
```json
{
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll": true,
    "source.organizeImports": true
  },
  "eslint.validate": [
    "javascript",
    "typescript",
    "svelte"
  ],
  "[svelte]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[javascript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "tailwindCSS.includeLanguages": {
    "svelte": "html"
  },
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.useLanguageServer": true,
  "files.eol": "\n"
}
```

### Additional Tips
- Use the **Command Palette** (`Ctrl+Shift+P`) to run formatting and linting commands.
- Enable **Auto Save** for a smoother workflow.
- Use the **REST Client** extension to test API endpoints directly from VS Code.
- Use **GitLens** for advanced Git history and code insights.

---

*Keep your editor and extensions up to date for the best experience!* 