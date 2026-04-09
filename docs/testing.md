# Testing

## Required layers

1. unit tests
2. contract tests
3. integration tests
4. end-to-end tests
5. recovery tests
6. approval-policy tests
7. chaos tests

## Required scenarios

- run planning from intake, bugfix, and resume specs
- workflow step sequencing for intake -> debate -> build -> test -> release
- workflow step sequencing for bugfix and resume paths
- task splitting
- implementation debate before coding
- build and test pass
- bugfix from reported issue
- approval required for risky changes
- resume from checkpoint after failure
- checkpoint selection for resume runs

## Safety checks

- block unsafe action attempts
- deny network beyond allowlist
- prevent destructive filesystem operations without approval
- keep secrets out of prompts/logs/artifacts
