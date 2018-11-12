import React from "react"
import PropTypes from "prop-types"
import { get } from "lodash"
import KeyValues from "../styled/KeyValueContainer"
import Tag from "../styled/Tag"
import Pre from "../styled/Pre"
import TagGroup from "../styled/TagGroup"

const TaskDefinitionSidebar = ({ data }) => {
  return (
    <KeyValues
      items={[
        { key: "Alias", value: get(data, "alias", "-") },
        { key: "Definition ID", value: get(data, "definition_id", "-") },
        { key: "Container Name", value: get(data, "container_name", "-") },
        { key: "Group Name", value: get(data, "group_name", "-") },
        { key: "Image", value: get(data, "image", "-") },
        {
          key: "Command",
          value: <Pre>{get(data, "command", "...")}</Pre>,
        },
        { key: "Memory", value: get(data, "memory", "-") },
        { key: "ARN", value: get(data, "arn", "-") },
        {
          key: "Tags",
          value: (
            <TagGroup>
              {get(data, "tags", [])
                .filter(tag => tag !== "")
                .map(tag => <Tag key={tag}>{tag}</Tag>)}
            </TagGroup>
          ),
        },
        {
          key: "Environment Vars",
          value: (
            <KeyValues
              items={get(data, "env", []).map((e, i) => ({
                key: e.name,
                value: <Pre>{e.value}</Pre>,
              }))}
            />
          ),
        },
      ]}
    />
  )
}

TaskDefinitionSidebar.displayName = "TaskDefinitionSidebar"
TaskDefinitionSidebar.propTypes = {
  data: PropTypes.object,
}

export default TaskDefinitionSidebar
