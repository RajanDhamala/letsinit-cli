# Contributing to LetsInit

Thank you for contributing to LetsInit. Contributions can add new starter
templates, improve existing templates, fix CLI behavior, or update documentation.

## Contribution workflow

1. Fork the [LetsInit CLI repository](https://github.com/RajanDhamala/letsinit-cli.git).
2. Clone your fork:

   ```bash
   git clone https://github.com/RajanDhamala/letsinit-cli.git
   cd cli
   ```

3. Create a branch for your work:

   ```bash
   git switch -c add-template-name
   ```

4. Make your changes and validate them locally.
5. Commit and push your branch:

   ```bash
   git status
   git add path/to/changed-files
   git commit -m "Add template name"
   git push -u origin add-template-name
   ```

6. Open a pull request from your branch to the main repository's `main` branch.
   Explain what changed, why it is useful, and how you tested it.

## Adding a new template

1. Create a clearly named directory under `templates/`.
2. Include only the files required to start and understand the project.
3. Add the template to `templateDirectories` and `stackChoices` in `cli.mjs`.
4. Add its dependency installation and start commands when they differ from the
   existing ecosystems.
5. Document the new stack in `README.md`.

Template `.gitignore` files must be stored as `_gitignore` so npm includes them
in the published package. LetsInit renames `_gitignore` to `.gitignore` when it
creates a project. If a template contains `.env.example`, LetsInit renames it to
`.env` in the generated project.

Do not commit real credentials, API keys, generated build output, dependency
directories, or local cache files.

## Updating an existing template

Keep changes focused on the template you are improving. Preserve its existing
language, project structure, package-manager support, and runtime conventions
unless the pull request intentionally changes them and explains why.

When changing dependencies, update the matching lockfile and verify that a newly
generated project installs successfully.

## Validation

Run the shared CLI checks:

```bash
node --check cli.mjs
git diff --check
npm pack --dry-run
```

Also generate the affected project through LetsInit and run the checks that fit
that stack, such as its build, lint, typecheck, tests, dependency installation,
or framework validation command.

## Pull request checklist

Before opening your pull request, confirm that:

- The change is focused and does not modify unrelated templates.
- New or updated templates are registered correctly in the CLI.
- No credentials, dependency directories, caches, or build output are included.
- Documentation reflects any visible setup or command changes.
- The shared CLI checks and relevant stack checks pass.
