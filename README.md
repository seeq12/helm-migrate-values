# helm-migrate-values

> **This is a Seeq-maintained fork** of
> [OctopusDeployLabs/helm-migrate-values](https://github.com/OctopusDeployLabs/helm-migrate-values),
> originally developed by Octopus Deploy. It has been modified from the upstream
> project — see [NOTICE](NOTICE) for a summary of the changes and
> [LICENSE](LICENSE) for terms. Octopus Deploy is not affiliated with this fork
> and does not support it.

A plugin to migrate user-specified Helm values between chart versions when the schema of `values.yaml` changes. Define migration paths with migration files in the chart repository to ensure seamless upgrades.
## Requirements

- **Helm** version 3 or newer
- **Git** if installing using the `helm plugin install` command below

## Install

```
$ helm plugin install https://github.com/seeq12/helm-migrate-values.git
```

## Usage

This Helm plugin enables users to migrate values across chart versions, accounting for changes in the values.yaml schema.

> **_NOTE:_** The intended usage of this plugin is that it does not apply any changes to the release on its own. It simply outputs the values to override, which are to be used with `helm upgrade --reset-then-reuse-values`.
>
> See [output values](#optional-output-the-migration-to-a-file) section for an example on how to apply the migration to a release

### Step 1: Define the Migration Files
Start by defining the migration files within your Helm chart. These files should be placed under the `value-migrations/` directory, relative to your base chart directory. You can customize the migration directory location by using the `--migration-dir` flag if necessary.

#### Migration File Naming Convention
Each migration file should conform to the following naming format:
`to-v{VERSION_TO}.yaml`, where **VERSION_TO** is the target chart's major version number. The plugin will use this file to define the transformation between the previous version and the specified version.

#### Migration File Structure
Migration files are written in YAML and use Go templating, similar to Helm templates. They leverage Sprig v3's [TxtFuncMap](https://github.com/Masterminds/sprig/blob/fc7fc0d6a0377bca7049c4a99e80b85f222d8caf/functions.go#L49) functions for transforming and mapping values between old and new schemas. See this [example](pkg/test-charts/v2/value-migrations/to-v2.yaml) of a migration definition from the integration test.

### Step 2: Run the Migration
To migrate your Helm release to a new chart version, use the following command:
```
helm migrate-values [RELEASE] [CHART] [flags]
```

- **RELEASE**: The name of the Helm release you're migrating.
- **CHART**: The chart you're migrating to, which can be a local chart (specified by file path) or a remote chart (using the `oci://` or `https://` prefixes).

## Example
```
helm migrate-values my-release oci://registry.example.com/charts/my-chart \
  --version 2.0.0 \
  -n my-namespace \
  --migration-dir my-chart/value-migrations \
  --output-file migrated-values.yaml
```

The `--output-file` flag allows you to optionally save the command's output to a file instead of displaying it in stdout, allow you to utilize it in subsequent Helm commands.

You can then use this file with the helm upgrade command to complete the migration:

```
helm upgrade [RELEASE] [CHART] -f migrated-values.yaml --reset-then-reuse-values
```

## Contributing

Please refer to the [Code of Conduct](CODE_OF_CONDUCT.md) before making any contributions.

We adhere to [Semantic Versioning](https://semver.org/).

Describe your change in the pull request itself. Release notes are written on the
GitHub release when one is cut, so there is no changelog file to update.
[CHANGELOG.md](CHANGELOG.md) is kept only as the record of releases up to 1.1.0.

## Releasing

Releases are cut from the GitHub Releases UI.

Before starting, `main` must already contain the release version in
`plugin.yaml`. This is the version `helm plugin list` reports, and the release
workflow refuses to publish if it does not match the tag. It lands through an
ordinary pull request. Then:

1. Go to **Releases -> Draft a new release**.
2. In the tag field, enter the new version and choose to create it on publish.
   Tags carry no `v` prefix, matching the existing tag history.
3. Write the release notes in the body. These are the changelog for this
   version -- the build will not overwrite them, and nothing else records them.
4. **Publish** the release. A draft does not create the tag, so nothing is built
   until you publish.

Publishing creates the tag, which starts the Release workflow. It checks the tag
against `plugin.yaml`, builds every platform, and uploads the archives and a
checksums file onto the release you just published. The release is live without
assets for the couple of minutes the build takes.

**Tags are immutable by organisation policy** -- they cannot be moved or deleted.
If the build fails after you have published, fix the problem and cut the next
patch version. The release can be deleted, but the tag cannot be reused.
