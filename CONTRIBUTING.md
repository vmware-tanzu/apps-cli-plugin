# Contributing to Apps Plugin for the Tanzu CLI

We welcome contributions from the community and first want to thank you for taking the time to contribute!

Please familiarize yourself with the [Code of Conduct](./CODE_OF_CONDUCT.md) before contributing.

* _CLA: Before you start working with Apps Plugin for the Tanzu CLI, please read and sign our Contributor License Agreement [CLA](https://cla.vmware.com/cla/1/preview). If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a Pull Request. For any questions about the CLA process, please refer to our [FAQ]([https://cla.vmware.com/faq](https://cla.vmware.com/faq))._

## Ways to contribute

We welcome many different types of contributions and not all of them need a Pull request. Contributions may include:

* New features and proposals
* Documentation
* Bug fixes
* Issue Triage
* Answering questions and giving feedback
* Helping to onboard new contributors
* Other related activities

## Getting started

### Development Environment Setup

Follow the documentation in the development document for the [Prerequisites](./DEVELOPMENT.md#Prerequisites) required for the setup, as well [Building](./DEVELOPMENT.md#building) the project.

### Run test

[How to run unit and e2e test](./DEVELOPMENT.md#testing)

Before a PR is accepted it will need to pass the all the checks on the CI.


## Contribution Flow

This project operates according to the talk, then code rule. If you plan to submit a pull request for anything more than a typo or obvious bug fix, first you should [raise an issue](https://github.com/vmware-tanzu/apps-cli-plugin/issues/new/choose) to discuss your proposal, before submitting any code.

This is a rough outline of what a contributor's workflow looks like:

* Make a fork of the repository within your GitHub account
* Create a topic branch in your fork from where you want to base your work
* Make commits of logical units
* Make sure your commit messages are with the proper format, quality and descriptiveness (see below)
* Push your changes to the topic branch in your fork
* Create a pull request containing that commit

We follow the GitHub workflow and you can find more details on the [GitHub flow documentation](https://docs.github.com/en/get-started/quickstart/github-flow).

Before submitting your pull request, we advise you to use the following:

### Pull Request Checklist

1. Check if your code changes will pass both code linting checks and unit tests.
2. Ensure all commits are signed using the -S (--gpg-sign) flag as below
```
git commit -S -m "<commit message>"
```
3. Ensure your commit messages are descriptive. We follow the conventions on [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/). Be sure to include any related GitHub issue references in the commit message. See [GFM syntax](https://guides.github.com/features/mastering-markdown/#GitHub-flavored-markdown) for referencing issues and commits.
4. Commit messages should follow the following format:

```
type: <short-summary>
```

Add ! after type if it's a breaking change
```
type!: <short-summary>
```

**NOTE**: short-summary should explain what the commit does. The automation tools we have built will take this short-summary and put it in a changelog. So write your short-summary in a clear and concise way.

`type`: can be one of the following:

- `build`: Changes that affect the build system (example: Makefile)
- `chore`: Housekeeping tasks, grunt tasks etc; no production code change
- `ci`: Changes to our CI configuration files and scripts
- `docs`: Documentation only changes
- `deprecate`: A code change that deprecates a feature
- `feat`: A new feature
- `fix`: A bug fix
- `perf`: A code change that improves performance
- `refactor`: A code change that neither fixes a bug nor adds a feature, includes updates to a feature.
- `removed`: Removed a feature that was previously deprecated
- `test`: Adding missing tests or correcting existing tests

## Reporting Bugs and Creating Issues

For specifics on what to include in your report, please follow the guidelines in the issue and pull request templates.

- [Pull Request Template](./.github/PULL_REQUEST_TEMPLATE.md)
- [Bug Report Template](./.github/ISSUE_TEMPLATE/bug_report.md)
- [Feature Request Template](./.github/ISSUE_TEMPLATE/feature_request.md)

## Ask for Help

The best way to reach us with a question when contributing is to ask on:

* The original GitHub issue

