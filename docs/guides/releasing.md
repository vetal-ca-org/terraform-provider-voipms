---
page_title: "Releasing a new version"
subcategory: ""
description: |-
  How to publish a new version of the VoIP.ms provider to the Terraform Registry and OpenTofu Registry.
---

# Releasing a new version

Public listings: [registry.terraform.io/providers/vetal-ca-org/voipms](https://registry.terraform.io/providers/vetal-ca-org/voipms) and [search.opentofu.org/provider/vetal-ca-org/voipms](https://search.opentofu.org/provider/vetal-ca-org/voipms/latest). Source address for both CLIs: `vetal-ca-org/voipms`.

GitHub Actions plus GoReleaser build the binaries, write checksums, and GPG-sign the checksums. That GitHub Release is the publish artifact for **both** registries. After each registry's first listing, a new tag is enough — do not click Publish again and do not re-upload the GPG key.

The Release workflow waits for Terraform Registry and OpenTofu Registry **in parallel** until each lists the tagged version.

## Prerequisites (one-time)

Already in place for this repository:

- Public GitHub repo `vetal-ca-org/terraform-provider-voipms`
- Namespace `vetal-ca-org` claimed in HCP Terraform, with signing key `F3ADF9C3A8C694B3`
- Actions secrets `GPG_PRIVATE_KEY` and `PASSPHRASE`
- Workflow [`.github/workflows/release.yml`](https://github.com/vetal-ca-org/terraform-provider-voipms/blob/master/.github/workflows/release.yml) and [`.goreleaser.yml`](https://github.com/vetal-ca-org/terraform-provider-voipms/blob/master/.goreleaser.yml)

### OpenTofu Registry (browser only)

OpenTofu does not accept CI, `gh`, or API submissions. Use the GitHub **issue form UI** once. Existing GitHub Releases (`v0.1.0`–current) are picked up after the listing is merged.

1. [Submit new Provider](https://github.com/opentofu/registry/issues/new?template=provider.yml). **Provider Repository:** `vetal-ca-org/terraform-provider-voipms`.
2. After that issue is merged, [Submit new Provider Signing Key](https://github.com/opentofu/registry/issues/new?template=provider_key.yml). **Provider Namespace:** `vetal-ca-org`. Paste `gpg --armor --export F3ADF9C3A8C694B3`. Your membership on `vetal-ca-org` must be [public](https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-personal-account-on-github/managing-your-membership-in-organizations/publicizing-or-hiding-organization-membership).

Until step 1 is merged, the OpenTofu Registry job fails immediately with HTTP 404. Re-run failed Release jobs after the listing exists.

## Each release

1. Land the change on `master` through a pull request. CI must be green (`make test` / Tests workflow).
2. From `master`, tag the **next** semantic version. There must not be a branch with the same name as the tag.

   | Change | Tag example |
   | --- | --- |
   | Bug fix or docs | `v0.1.1` |
   | New resource / data source, compatible | `v0.2.0` |
   | Breaking schema or behavior | `v1.0.0` |

   ```shell
   git checkout master
   git pull
   git tag v0.1.1
   git push origin v0.1.1
   ```

3. Confirm **Actions → Release** succeeds:

   - **goreleaser** publishes platform zips, `terraform-provider-voipms_<version>_manifest.json`, `_SHA256SUMS`, and `_SHA256SUMS.sig`.
   - **Terraform Registry** and **OpenTofu Registry** run concurrently and wait until each listing includes the version (OpenTofu's crawler is typically a 15-minute cycle, so that job may take longer).

4. Consumers upgrade with `terraform init -upgrade` or `tofu init -upgrade` (or a tighter `version` constraint).

## Do not

- Reuse or move an existing tag (`v0.1.0` stays `v0.1.0` forever).
- Replace zip files or checksums on a published GitHub Release.
- Remove the GPG public key from the `vetal-ca-org` namespace. To rotate, **add** a new key, update the Actions secrets, then tag; leave the old key so older versions still verify. Submit the new public key to OpenTofu the same way as the first key.
- Put secrets, live phone numbers, or SIP passwords in git. Docs and examples use fictional `555` numbers.
- Open OpenTofu registry issues with `gh` or the GitHub API. Those are closed unprocessed.

## Registry docs

Provider pages on both registries come from `docs/` in the tagged commit. After you change examples or templates, regenerate before merging:

```shell
make generate
```

The household-style configuration on the provider page is [`examples/complete/main.tf`](https://github.com/vetal-ca-org/terraform-provider-voipms/blob/master/examples/complete/main.tf).
