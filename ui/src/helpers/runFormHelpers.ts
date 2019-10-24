import { FieldSpec } from "../types"

export const clusterFieldSpec: FieldSpec = {
  name: "cluster",
  label: "Cluster",
  description: "Select a cluster for this task to execute on.",
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
