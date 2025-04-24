# Releasing

## Release Process

This release process covers the steps to release new major and minor versions for the `opentelemetry-collector` with Kyma-specific customizations.

1. Verify that all issues in the [GitHub milestone](https://github.com/kyma-project/opentelemetry-collector-components/milestones) related to the version are closed.
2. Close the milestone.

3. Create a new [GitHub milestone](https://github.com/kyma-project/opentelemetry-collector-components/milestones) for the next version.

4. Manually trigger the [Create release](https://github.com/kyma-project/opentelemetry-collector-components/actions/workflows/create-release.yaml) GitHub Action from the target branch and provide an appropriate release tag as an arugment. The target branch must be:
  - `main` for major and minor releases,
  - `release-x.y` for patch releases of version x.y.

> [!NOTE]
> - If you check `Do-Not-Publish`, the release will not be published. However, the action still creates an image and a draft release.
> - If you check `Force`, it overwrites the existing release and any previously built images. It will never overwrite an already published release. For draft releases, the force option: - recreates the Git tag in the repository - recreates the release notes - overwrites the previously built image

5. If the previous release was a bugfix version (patch release) that contains cherry-picked changes, these changes might appear again in the generated change log. If there are redundant entries, edit the release description and remove them.

## Changelog

Every PR's title must adhere to the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification for an automatic changelog generation. It is enforced by a [semantic-pull-request](https://github.com/marketplace/actions/semantic-pull-request) GitHub Action.

### Pull Request Title

Because of the squash-and-merge GitHub workflow, each PR results in a single commit after merging into the main development branch. The PR's title becomes the commit message and must adhere to the template:

`type(scope?): subject`

#### Type

- **feat**: A new feature or functionality change.
- **fix**: A bug or regression fix.
- **docs**: Changes regarding the documentation.
- **test**: The test suite alternations.
- **deps**: The changes in the external dependencies.
- **chore**: Anything not covered by the above categories (such as refactoring or artefacts building alternations).

Note that PRs of type `chore` do not appear in the change log for the release. Therefore, exclude maintenance changes that are not interesting to consumers of the project by marking them as "chore", for example:

- Dotfile changes (.gitignore, .github, and so forth).
- Changes to development-only dependencies.
- Minor code style changes.
- Formatting changes in documentation.

#### Subject

The subject must describe the change and follow the recommendations:

- Describe a change using the [imperative mood](https://en.wikipedia.org/wiki/Imperative_mood).
 It must start with a present-tense verb, for example (but not limited to) Add, Document, Fix, Deprecate.
- Start with an uppercase, and not finish with a full stop.
- Kyma [capitalization](https://github.com/kyma-project/community/blob/main/docs/guidelines/content-guidelines/02-style-and-terminology.md#capitalization) and [terminology](https://github.com/kyma-project/community/blob/main/docs/guidelines/content-guidelines/02-style-and-terminology.md#terminology) guides.
