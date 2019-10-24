import { FieldSpec } from "../types"

export const aliasFieldSpec: FieldSpec = {
  name: "alias",
  label: "alias",
  description: "alias",
  initialValue: "",
}
export const groupNameFieldSpec: FieldSpec = {
  name: "group_name",
  label: "Group Name",
  description:
    "Create a new group name or select an existing one to help searching for this task in the future.",
  initialValue: "",
}
export const imageFieldSpec: FieldSpec = {
  name: "image",
  label: "Docker Image",
  description: "The full URL of the Docker image and tag.",
  initialValue: "",
}
export const commandFieldSpec: FieldSpec = {
  name: "command",
  label: "Command",
  description: "The command for this task to execute.",
  initialValue: "",
}
export const memoryFieldSpec: FieldSpec = {
  name: "memory",
  label: "Memory (MB)",
  description: "The amount of memory (MB) this task needs.",
  initialValue: 1024,
}
export const cpuFieldSpec: FieldSpec = {
  name: "cpu",
  label: "CPU (Units)",
  description:
    "The amount of CPU (units) this task needs. Note: 1 CPU unit is approximately equivalent to 1 MB.",
  initialValue: 512,
}
export const tagsFieldSpec: FieldSpec = {
  name: "tags",
  label: "Tags",
  description: "",
  initialValue: [],
}
export const envFieldSpec: FieldSpec = {
  name: "env",
  label: "Environment Variables",
  description: "",
  initialValue: [],
}
