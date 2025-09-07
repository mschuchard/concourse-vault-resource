### 1.2.3
- Improve metadata formatting.
- Improve credential generation and expiratiom time calculation.
- Validate raw secret before metadata conversion.
- Improve error logging.
- Fix incorrect returned version of dynamic secret post-renewal.
- Fix validation of secret renewal in `check` step.

### 1.2.2
- Improve expiration time format in metadata version.
- Properly utilize `aws_mount_path` `source` parameter.
- Improve Vault client initialization efficiency.

### 1.2.1
- Various code optimization and improvements.

### 1.2.0
- Remove `renew` parameter from in/get step.
- Add Azure, Consul, Kubernetes, Nomad, RabbitMQ, SSH, and Terraform Cloud secret engine generate credentials support.

### 1.1.1
- Improve Lease ID validation.
- Warn instead of error if a single secret interaction fails.
- Do not attempt to append secret metadata if secret interaction fails.

### 1.1.0
- `check` step automatic renewal of dynamic secrets (beta).

### 1.0.2
- Gracefully handle errors instead of simple logging and panic.
- Fix error collection for Vault read/write interactions during in/out steps.

### 1.0.1
- Improve Vault config validation and defaults.
- Collect all errors during Vault interactions for logging.

### 1.0.0
- Initial release.
