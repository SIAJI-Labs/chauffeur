# Contributing to Chauffeur

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification for our commit messages. This leads to more readable messages that are easy to follow when looking through the project history.

### Commit Message Format
Each commit message consists of a **header**, a **body** and a **footer**. The header has a special format that includes a **type**, a **scope** and a **subject**:

```
<type>(<scope>): <subject>
<BLANK LINE>
<body>
<BLANK LINE>
<footer>
```

### Types
Must be one of the following:

* **feat**: A new feature
* **fix**: A bug fix
* **docs**: Documentation only changes
* **style**: Changes that do not affect the meaning of the code (white-space, formatting, etc)
* **refactor**: A code change that neither fixes a bug nor adds a feature
* **perf**: A code change that improves performance
* **test**: Adding missing tests or correcting existing tests
* **chore**: Changes to the build process or auxiliary tools and libraries such as documentation generation

### Scope
The scope should be the name of the module affected (as perceived by the person reading the changelog generated from commit messages).

### Subject
The subject contains a succinct description of the change:

* use the imperative, present tense: "change" not "changed" nor "changes"
* don't capitalize the first letter
* no dot (.) at the end

### Examples

```
feat(cli): add service status command

Add new command to check the status of installed services.
The command provides detailed information about service state and uptime.

Closes #123
```

```
fix(installer): resolve nginx config template issue

Fixed incorrect template variable substitution that caused nginx 
configuration generation to fail on certain systems.

Fixes #456
```

```
docs(readme): update installation instructions

Update the README with more detailed installation steps and 
troubleshooting information.
```