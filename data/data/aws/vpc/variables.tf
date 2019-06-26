variable "availability_zones" {
  type        = list(string)
  description = "The availability zones in which to provision subnets."
}

variable "cidr_block" {
  type = string
}

variable "cluster_id" {
  type = string
}

variable "private_master_endpoints" {
  description = "If set to true, private-facing ingress resources are created."
  default     = true
}

variable "public_master_endpoints" {
  description = "If set to true, public-facing ingress resources are created."
  default     = true
}

variable "region" {
  type        = string
  description = "The target AWS region for the cluster."
}

variable "tags" {
  type        = map(string)
  default     = {}
  description = "AWS tags to be applied to created resources."
}

variable "open_vxlan_ports" {
  type = bool

  description = "For OCP 4.2 only! Causes VXLAN ports to be opened."
}

variable "open_ovn_ports" {
  type = bool

  description = "For OCP 4.2 only! Causes ovn-kubernetes-related ports to be opened."
}
