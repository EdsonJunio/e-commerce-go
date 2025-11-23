# 🛡️ Quality & Contribution Standards

This project uses **Git Hooks** to ensure high quality, consistency, and security throughout the development workflow.

---

## 🚀 1. Quick Setup (First Time Only)

After cloning the repository, run the installation script to enable git hooks:

```bash
./scripts/install-hooks.sh
Without this step, automatic validations will not run.

📏 2. Commit Messages — Conventional Commits
We follow the Conventional Commits specification:

php-template
Copiar código
<type>(<scope>): <subject>
Accepted Types
Type	Description	Example
feat	New feature	feat(auth): add login endpoint
fix	Bug fix	fix(cart): calculate total correctly
chore	Maintenance / configuration	chore: update go.mod
docs	Documentation	docs(readme): update setup guide
test	Tests	test(user): add unit tests
refactor	Code refactoring	refactor: simplify loop

⚠️ Commits that do not follow this pattern will be automatically rejected by the commit-msg hook.

🌿 3. Branch Strategy
We use a simplified Git Flow approach.

Protected Branches
main

develop

Direct pushes to these branches are strictly prohibited.

Allowed Patterns
php-template
Copiar código
feature/<name>
bugfix/<name>
hotfix/<name>
Valid Examples
feature/user-login

bugfix/payment-error

Invalid Examples
my-test

login-fix

Recommended Workflow
Create branch

Develop

Commit

Push

Open Pull Request

🧪 4. Automated Testing Pipeline
Before every commit, the pre-commit hook runs:

go test ./... (unit tests)

coverage calculation

(optional) linting and formatting

If any test fails, the commit will be blocked.

🆘 Troubleshooting
In emergency situations, you may bypass the validations:

bash
Copiar código
git commit -m "message" --no-verify
git push origin my-branch --no-verify
Only use this if you fully understand the implications.