# Releasing the Terraform provider

1. Merge all your changes to the stable branch
1. If any of the changes upgrades Go version, make sure that you have followed the [Go upgrade checklist](https://github.com/UpCloudLtd/terraform-provider-upcloud/blob/v2.4.1/DEVELOPING.md#go-version-upgrades).
1. Update CHANGELOG.md
    1. Add new heading with the correct version e.g. `## [v2.3.5]`
    1. Update links at the bottom of the page
    1. Leave "Unreleased" section at the top empty
1. Tag the release `vX.Y.Z` (e.g. `v2.3.5`)
1. Push the tag to GitHub
1. Ensure that there is new [release in GitHub](https://github.com/UpCloudLtd/terraform-provider-upcloud/releases) and publish it if it is draft state
1. The [Terraform registry](https://registry.terraform.io/providers/UpCloudLtd/upcloud) will pick up the release
