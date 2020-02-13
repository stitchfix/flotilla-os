import { FormikActions } from "formik"
import { createMemoryHistory, createLocation } from "history"
import { RouteComponentProps } from "react-router-dom"
import {
  Task,
  Run,
  RunStatus,
  ExecutionEngine,
  NodeLifecycle,
  ExecutableType,
} from "../types"

export function createMockRouteComponentProps<MatchParams>({
  path,
  url,
  params,
}: {
  path: string
  url: string
  params: MatchParams
}): RouteComponentProps {
  return {
    history: createMemoryHistory(),
    match: {
      isExact: false,
      path,
      url,
      params,
    },
    location: createLocation(url),
  }
}

export const mockFormikActions: FormikActions<any> = {
  setStatus: jest.fn(),
  setError: jest.fn(),
  setErrors: jest.fn(),
  setSubmitting: jest.fn(),
  setTouched: jest.fn(),
  setValues: jest.fn(),
  setFieldValue: jest.fn(),
  setFieldError: jest.fn(),
  setFieldTouched: jest.fn(),
  validateForm: jest.fn(),
  validateField: jest.fn(),
  resetForm: jest.fn(),
  submitForm: jest.fn(),
  setFormikState: jest.fn(),
}

export const createMockTaskObject = (overrides?: Partial<Task>): Task => ({
  env: [{ name: "a", value: "b" }],
  arn: "arn",
  definition_id: "my_definition_id",
  image: "image",
  group_name: "group_name",
  container_name: "container_name",
  alias: "alias",
  memory: 1024,
  cpu: 512,
  command: "command",
  tags: ["a", "b", "c"],
  privileged: false,
  gpu: 0,
  adaptive_resource_allocation: true,
  ...overrides,
})

export const createMockRunObject = (overrides?: Partial<Run>): Run => ({
  instance: {
    dns_name: "my_dns_name",
    instance_id: "my_instance_id",
  },
  task_arn: "my_task_arn",
  run_id: "my_run_id",
  definition_id: "my_definition_id",
  alias: "my_alias",
  image: "my_image",
  cluster: "my_cluster",
  status: RunStatus.RUNNING,
  started_at: "2019-10-24T05:21:51",
  group_name: "group_name",
  env: [],
  cpu: 1,
  memory: 1024,
  command: "echo 'hi'",
  queued_at: "queued_at",
  engine: ExecutionEngine.ECS,
  node_lifecycle: NodeLifecycle.SPOT,
  max_cpu_used: 1,
  max_memory_used: 1,
  pod_name: "pod",
  cloudtrail_notifications: {},
  executable_id: "my_executable_id",
  executable_type: ExecutableType.ExecutableTypeDefinition,
  execution_request_custom: {},
  ...overrides,
})
