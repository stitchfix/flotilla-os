export default (definitionId = "definitionId") => ({
  env: [
    {
      name: "key_a",
      value: "value_a",
    },
    {
      name: "key_b",
      value: "value_b",
    },
  ],
  arn: "my_arn",
  definition_id: definitionId,
  image: "image_repo/image_name:image_tag",
  group_name: "my_group",
  container_name: "my_container",
  user: "nobody",
  alias: "my_alias",
  memory: 500,
  command: "echo 'hi'",
  tags: ["tag_a", "tag_b"],
})
