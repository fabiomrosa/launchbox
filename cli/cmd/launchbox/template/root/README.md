<!-- INITIATE VARS --->

[PRD__URL]: https://<DOMAIN>.com
[STG__URL]: https://<STAGING-DOMAIN>
[NODE__URL]: https://nodejs.org/en/download
[PNPM__URL]: http://pnpm.io/installation
[DOCKER__URL]: https://docs.docker.com/get-started
[NEXT__URL]: https://nextjs.org/docs
[REACT__URL]: https://react.dev/reference/react
[TYPESCRIPT__URL]: https://www.typescriptlang.org/docs
[COMMIT_STYLEGUIDE__URL]: https://www.conventionalcommits.org/en/v1.0.0
[PRD__BADGE]: https://img.shields.io/badge/Production-00AB68?style=flat&logo=
[STG__BADGE]: https://img.shields.io/badge/Staging-00AB68?style=flat&logo=
[GITHUB__BADGE]: https://img.shields.io/badge/Github-000?style=flat&logo=github
[NEXT__BADGE]: https://img.shields.io/badge/Next-000?style=flat&logo=vercel
[REACT__BADGE]: https://img.shields.io/badge/React-fff?style=flat&logo=react
[TYPESCRIPT__BADGE]: https://img.shields.io/badge/Typescript-eee?style=flat&logo=typescript
[PNPM__BADGE]: https://img.shields.io/badge/pnpm-F69220?logo=pnpm&logoColor=fff
[DOCKER__BADGE]: https://img.shields.io/badge/Docker-2496ED?logo=docker&logoColor=fff

<!-- END VARS --->

# <PROJECT_NAME>

[![production][PRD__BADGE]][PRD__URL]
[![staging][STG__BADGE]][STG__URL]

> Add Project Description

## Tech references

[![next][NEXT__BADGE]][NEXT__URL]
[![react][REACT__BADGE]][REACT__URL]
[![typescript][TYPESCRIPT__BADGE]][TYPESCRIPT__URL]
[![pnpm][PNPM__BADGE]][PNPM__URL]
[![docker][DOCKER__BADGE]][DOCKER__URL]

## Requirements

To run this project on your computer, you need to install a few dependencies:

- [Node.js][NODE__URL] 24.x or above
- [PnPM][PNPM__URL] 10.x or above (optional if using Docker)
- [Docker][DOCKER__URL] (optional if using PnPM)

## Running locally

Clone the project:

```bash
git clone git@github.com:<ORG>/<project_name>.git
```

Go to the project directory:

```bash
cd <project_name>
```

Install dependencies:

```bash
pnpm install
```

Start the development server:

```bash
pnpm dev
```

## Config

In the project root folder:

```bash
cp .env.example .env.local
```

For each `key`, add the corresponding value, like:

- `NEXT_PUBLIC_BACKEND_API_V1=https://backend.<project_name>.com`

> WARN: Perform this step only if the project uses environment variables.

## Standards

We follow two simple standards to improve organization and increase the maintainability of our project, keeping our commits and branches standardized.

We expect you to follow them too 😍 … seriously, please follow these standards 🙏

### Branches

Our branch naming convention is very simple — no fuss:

Branch examples following this pattern, using the `kebab-case` convention:
`<type>/issue-<number>/<verb>-<description>`

**Type** — can be one of the following:

```
  chore: configuration files
  docs: documentation changes
  feat: a new feature
  fix: a bug fix
  perf: code change focused on improving performance
  refactor: code change that neither adds a feature nor fixes a bug
  test: adding or fixing tests
```

**Number** — the issue number from the opened issue.

**Verb** — can be one of the following:

```
  add: for new features
  change: for changes to existing functionality
  remove: for removed features
```

Examples:

- `docs/issue-42/add-readme`
- `chore/issue-42/change-googlemaps-deps`
- `feat/issue-42/add-user-auth`
- `fix/issue-42/remove-modal`

### Commits

For commit messages, we follow the [Conventional Commits][COMMIT_STYLEGUIDE__URL] standard, enforced by `commitlint`.

Commits examples following this pattern:
`<type>(<scope>?): [<verb> <description>;]`

**Type** — can be one of the following:

```
  build: changes that affect the build system or external dependencies. (e.g., pnpm, ESLint, Husky, etc.)
  chore: configuration-only files (e.g., commitlint.config.mjs, docker-compose.yml, .editorconfig, .gitignore, etc.)
  ci: changes to CI/CD configuration files and scripts (e.g., GitHub Actions, CircleCI, Travis, etc.)
  docs: documentation-only changes (e.g., README.md, CHANGELOG.md, CONTRIBUTING.md, etc.)
  feat: a new feature
  fix: a bug fix
  perf: code change focused on improving performance
  refactor: code change that neither adds a feature nor fixes a bug
  revert: revert a previous commit
  style: code changes that do not affect the meaning of the code (e.g., whitespace, formatting, semicolons, etc.)
  test: adding or fixing tests
```

**Scope** — MUST be a noun describing a section of the codebase, written inside parentheses. (e.g., fix(parser):)

**Verb** — MUST be in the imperative form. (e.g., add, fix, remove, rename, update, etc.)

**Description** — A short description of the change. Multiple descriptions may be included, separated by semicolons.

Examples:

- `chore: drop support for Node v18`
- `docs(CHANGELOG): change date in last release`
- `feat(auth): add login endpoint; update password reset flow`
- `fix(home): remove modal`
