# Concourse Vault Resource

A [concourse-ci](https://concourse-ci.org) resource for interacting with secrets via [Vault](https://www.vaultproject.io). This plugin's code at `HEAD` is currently tested against Vault version 1.14.10. The most recent release was tested against 1.14.10.

This resource's container image is currently hosted at [matthewschuchard/concourse-vault-resource](https://hub.docker.com/repository/docker/matthewschuchard/concourse-vault-resource) for usage within Concourse.

This repository and project is based on the work performed for [MITODL](https://github.com/mitodl/concourse-vault-resource), and now serves as an upstream for the project hosted within that organization. Accordingly it maintains the BSD-3 license with copyright notice.

## Behavior

### `source`: designates the Vault server and authentication engine information

**parameters**
- `auth_engine`: _optional_ The authentication engine for use with Vault. Allowed values are `aws` or `token`. If unspecified will default to `aws` with no `token` parameter specified, or `token` if `token` parameter is specified.

- `address`: _optional_ The address for the Vault server in format of `URL:PORT`. default: `http://127.0.0.1:8200`

- `aws_mount_path`: _optional_ The mount path for the AWS authentication engine. Parameter is ignored if authentication engine is not `aws`. default: `aws`

- `aws_vault_role`: _optional_ The Vault role for the AWS authentication login to Vault. Parameter is ignored if authentication engine is not `aws`. default: (Vault role in utilized AWS authentication engine with the same name as the current utilized AWS IAM Role)

- `token`: _optional_ The token for the token authentication engine. Required if `auth_engine` parameter is `token`.

- `insecure`: _optional_ Whether to utilize an insecure connection with Vault (e.g. no HTTP or HTTPS with self-signed cert). default: `false`

- `secret`: _required/optional_ Required for `check` step. Mutually exclusive with `params` for `in` step, but one of the two must be specified. Note this value is ignored during `out` as it is not possible for it to have any effect with that step's functionality. The following YAML schema is required for the secret specification.

```yaml
secret:
  engine: <secret engine>
  mount: <secret mount path>
  path: <secret path>
```

### `version`: designates the specific version of a secret

NOTES:
- The KV1 secret engine does not support versioning.
- The KV2 secret engine currently returns the latest version of a secret if version is input as `"0"`, but this behavior may be subject to changes in the API, and no version should be specified if the latest is desired.
- The `version` input is ignored for `in` with `params` as it is associated with a single secret path, and therefore only functions when peered with `source` for `check` or `in`.

**parameters**
- `version`: _optional_ The following YAML schema is required for the version specification.

```yaml
version:
  version: <version>
```

Note that the response `version` schema for the `in` and `out` steps is different because multiple secrets can be specified, and for those steps' responses `version` is descriptive rather than functional. Therefore the version information displayed in Concourse for those steps appears like:

```yaml
version:
  <mount>-<path>: <version>
```

### `check`: returns secret versions between input version and retrieved version sequentially and inclusive

NOTE: currently only the KV2 secrets engine is supported.  
NOTE: if the specified secret is dynamic, then the input version is ignored because the comparison is between the current time and the secret expiration time

This step has no parameters, and utilizes the `source` and `version` values for functionality. It also executes automatically during resource instantiation.

Example output for a KV2 secret with Concourse input version `1` and retrieved Vault version `3`:

```json
[{"version":"1"},{"version":"2"},{"version":"3"}]
```

### `in`: interacts with the supported Vault secrets engines to retrieve and generate secrets

**parameters**

NOTE: For dynamic secret renewal the `path` must be suffixed with the same `/` and SHA suffix from its associated Lease ID (this can be inspected within the Concourse metadata returned when generating the initial dynamic secret).

- `<secret_mount path>`: _required/optional_ Mutually exclusive with `source.secret`, but one of the two must be specified. One or more map/hash/dictionary of the following YAML schema for specifying the secrets to retrieve, generate, or renew.

```yaml
<secret_mount_path>:
  paths:
  - <path/to/secret>
  - <path/to/other_secret>
  engine: <secret engine> # supported values: database, aws, kv1, kv2
  renew: false # whether to renew the dynamic secret(s) instead of generating
```

**usage**

The retrieved secrets and their associated values are written/appended as JSON formatted strings to a file located at `/opt/resource/vault.json` for subsequent loading and parsing in the pipeline with the following schema:

```yaml
---
<MOUNT>-<PATH>: <SECRET VALUES>
```

```json
{ "<MOUNT>-<PATH>": <SECRET VALUES> }
```

Examples:

```yaml
---
secret-foo/bar:
  password: supersecret
```

```json
{
  "secret-foo/bar": {
    "password": "supersecret"
  }
}
```

### `out`: interacts with the supported Vault secrets engines to populate secrets

- `<secret_mount path>`: _required_ One or more map/hash/dictionary of the following YAML schema for specifying the secrets to populate.

```yaml
<secret_mount_path>:
  secrets:
    <path/to/secret>:
      <key>: <value>
    <path/to/other_secret>:
      <key>: <value>
      <key>: <value>
  engine: <secret engine> # supported values: kv1, kv2
  patch: <boolean> # default: false; also see below
```

Although optimally `patch` would be specified per path, this would be cumbersome in both implementation and usage, and therefore it is specified for all paths for a given `mount`. When `patch` is specified as `true`, then (from [Vault API PKG documentation](https://pkg.go.dev/github.com/hashicorp/vault/api#KVv2.Patch)):

> Patch additively updates the most recent version of a key-value secret, differentiating it from Put which will fully overwrite the previous data. Only the key-value pairs that are new or changing need to be provided.

The default value of `false` will trigger the `Put` behavior of overwriting/replacing all values at the specified secret path. **Note that the `patch` nested parameter only functions if the engine is kv2, and is ignored if the engine is kv1.**

## Example

```yaml
resource_types:
- name: vault
  type: docker-image
  source:
    repository: matthewschuchard/concourse-vault-resource:1.0
    tag: latest

resources:
- name: vault
  type: vault
  source:
    address: https://mitodl.vault.com:8200
    auth_engine: aws
- name: vault-secret-check
  type: vault
  source:
    address: https://mitodl.vault.com:8200
    token: abcdefghijklmnopqrstuvwxyz09
    secret:
      engine: kv2
      mount: secret
      path: path/to/secret

jobs:
- name: do something
  plan:
  - get: my-code
  - get: vault
    params:
      database-mitxonline:
        paths:
        - readonly
        - other_readonly
        engine: database
      database-two:
        paths:
        - admin/abcdefghijk123456789lmno
        engine: database
        renew: true
      secret:
        paths:
        - path/to/secret
        engine: kv2
      kv:
        paths:
        - path/to/secret
        engine: kv1
  - put: vault
    params:
      secret:
        secrets:
          path/to/secret:
            key: value
            other_key: other_value
        engine: kv2
        patch: true
      kv:
        secrets:
          path/to/secret:
            key: value
          path/to/other_secret:
            key: value
        engine: kv1
  - get: vault-secret-check
```

## Contributing
Code should pass all unit and acceptance tests. New features should involve new unit tests.

Please consult the GitHub Project for the current development roadmap.
