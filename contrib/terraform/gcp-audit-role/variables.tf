/*
 * Copyright (c) 2026 TrustReady Inc <hello@trustready.io>.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

variable "trustready_issuer_url" {
  type        = string
  description = <<-EOT
    The issuer URL TrustReady mints its assertions under, unique to your TrustReady
    organization. Copy it exactly: GCP compares it case-sensitively and the
    last path segment is a mixed-case identifier.
  EOT

  validation {
    condition     = can(regex("^https://[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?(/[A-Za-z0-9._~%-]+)*/?$", var.trustready_issuer_url))
    error_message = "The issuer must be an https:// URL with no port and no query string."
  }
}

variable "trustready_subject" {
  type        = string
  description = <<-EOT
    The subject claim TrustReady asserts, identifying your TrustReady organization. The
    provider trusts this value and no other.
  EOT

  validation {
    condition     = can(regex("^[A-Za-z0-9_-]+$", var.trustready_subject))
    error_message = "The subject must be the identifier TrustReady showed you."
  }
}

variable "service_account_name" {
  type        = string
  default     = "trustready-audit"
  description = <<-EOT
    Account id of the service account TrustReady impersonates. Use the same name
    in every project, and record the email on the connector if you change it.
  EOT

  validation {
    condition     = can(regex("^[a-z]([-a-z0-9]{4,28}[a-z0-9])$", var.service_account_name))
    error_message = "The service account name must be a valid GCP service account id (6-30 characters)."
  }
}

variable "pool_id" {
  type        = string
  default     = "trustready"
  description = <<-EOT
    Workload Identity pool id. Change it only when this project already has
    a pool with this id.
  EOT

  validation {
    condition     = can(regex("^[a-z0-9](?:[a-z0-9-]{2,30}[a-z0-9])$", var.pool_id)) && !startswith(var.pool_id, "gcp-")
    error_message = "The pool id must be 4-32 characters of [a-z0-9-] and must not start with gcp-."
  }
}

variable "provider_id" {
  type        = string
  default     = "trustready"
  description = <<-EOT
    OIDC provider id inside the pool. Change it only when this pool already
    has a provider with this id.
  EOT

  validation {
    condition     = can(regex("^[a-z0-9](?:[a-z0-9-]{2,30}[a-z0-9])$", var.provider_id)) && !startswith(var.provider_id, "gcp-")
    error_message = "The provider id must be 4-32 characters of [a-z0-9-] and must not start with gcp-."
  }
}

variable "required_bucket_location" {
  type        = string
  default     = "global"
  description = <<-EOT
    Location of the project's _Required log bucket. Leave the default unless
    Cloud Logging default resource settings moved that bucket. TrustReady queries
    the same location; a mismatch leaves last login unknown.
  EOT

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]*$", var.required_bucket_location))
    error_message = "The _Required bucket location must be a Cloud Logging location such as global or us-central1."
  }
}
