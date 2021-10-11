# Releasing the Terraform provider

1. Merge all your changes to the stable branch
1. Update CHANGELOG.md
    1. Add new heading with the correct version e.g. `## [v2.3.5]`
    1. Update links at the bottom of the page
    1. Leave “Unreleased” section at the top empty
1. Tag the release `vX.Y.Z` (e.g. `v2.3.5`)
1. Push the tag to GitHub
1. Push the tag to the internal GitLab repo. This will fire a CI job that signs the code & create a new draft release for the tag
1. Open the [draft release in GitHub](https://github.com/UpCloudLtd/terraform-provider-upcloud/releases) & publish it when you are happy with it
1. The [Terraform registry](https://registry.terraform.io/providers/UpCloudLtd/upcloud) will pick up the release
