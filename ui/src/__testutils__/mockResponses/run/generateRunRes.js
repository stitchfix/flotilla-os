export default runId => ({
  instance: {
    dns_name: "some_dns_name",
    instance_id: "some_instance_id",
  },
  task_arn: "some_task_arn",
  run_id: runId,
  definition_id: "some_definition_id",
  cluster: "some_cluster",
  exit_code: 0,
  status: "some_status",
  started_at: "some_started_at",
  finished_at: "some_finished_at",
  group_name: "some_group_name",
  env: [
    {
      name: "foo",
      value: "bar",
    },
  ],
})
