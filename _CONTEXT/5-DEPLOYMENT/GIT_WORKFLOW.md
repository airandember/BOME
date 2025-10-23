# BOME Git Workflow

## Branching Strategy

### Main Branches
- **main** - Production-ready code
- **develop** - Integration branch for features

### Feature Branches
- **feature/backend-setup** - Backend development
- **feature/frontend-setup** - Frontend development
- **feature/roku-app** - Roku app development
- **feature/admin-dashboard** - Admin dashboard development
- **feature/integration** - API integration and testing

### Naming Convention
- Feature branches: `feature/description`
- Bug fixes: `fix/description`
- Hotfixes: `hotfix/description`
- Releases: `release/version`

## Development Workflow

1. **Create Feature Branch**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/new-feature
   ```

2. **Development**
   - Make changes and commit frequently
   - Use descriptive commit messages
   - Keep commits atomic and focused

3. **Push and Create Pull Request**
   ```bash
   git push origin feature/new-feature
   # Create PR to develop branch
   ```

4. **Code Review**
   - Review code changes
   - Address feedback
   - Update branch if needed

5. **Merge to Develop**
   - Merge feature branch to develop
   - Delete feature branch

6. **Release to Main**
   - Create release branch from develop
   - Test and fix any issues
   - Merge to main and tag release

## Commit Message Format
```
type(scope): description

[optional body]

[optional footer]
```

### Types
- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes
- **refactor**: Code refactoring
- **test**: Adding tests
- **chore**: Maintenance tasks

### Examples
```
feat(backend): add user authentication API
fix(frontend): resolve video player loading issue
docs(readme): update installation instructions
```

## Branch Protection Rules

### Main Branch
- Require pull request reviews
- Require status checks to pass
- Require branches to be up to date
- Restrict direct pushes

### Develop Branch
- Require pull request reviews
- Require status checks to pass
- Allow force pushes for maintainers

## Release Process

1. Create release branch from develop
2. Update version numbers
3. Update changelog
4. Test thoroughly
5. Create pull request to main
6. Merge and tag release
7. Deploy to production
8. Merge back to develop

## Emergency Hotfixes

1. Create hotfix branch from main
2. Fix the issue
3. Test thoroughly
4. Create pull request to main
5. Merge and tag hotfix release
6. Deploy immediately
7. Merge back to develop 