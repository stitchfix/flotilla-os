import React, { Fragment } from "react"
import PropTypes from "prop-types"
import { get, omit } from "lodash"
import KeyValues from "../styled/KeyValues"
import Tag from "../styled/Tag"
import Pre from "../styled/Pre"
import TagGroup from "../styled/TagGroup"

const TaskDefinitionSidebar = ({ data }) => {
  return (
    <Fragment>
      <KeyValues
        raw={omit(data, "env")}
        label="Task Definition"
        items={{
          Alias: get(data, "alias", "-"),
          "Definition ID": get(data, "definition_id", "-"),
          "Container Name": get(data, "container_name", "-"),
          "Group Name": get(data, "group_name", "-"),
          Image: get(data, "image", "-"),
          Command: <Pre>{get(data, "command", "...")}</Pre>,
          Memory: get(data, "memory", "-"),
          ARN: get(data, "arn", "-"),
          Tags: (
            <TagGroup>
              {get(data, "tags", [])
                .filter(tag => tag !== "")
                .map(tag => <Tag key={tag}>{tag}</Tag>)}
            </TagGroup>
          ),
        }}
      />
      <KeyValues
        raw={get(data, "env")}
        label="Environment Variables"
        items={get(data, "env", []).reduce((acc, env) => {
          acc[env.name] = <Pre>{env.value}</Pre>
          return acc
        }, {})}
      />
    </Fragment>
  )
}

TaskDefinitionSidebar.displayName = "TaskDefinitionSidebar"
TaskDefinitionSidebar.propTypes = {
  data: PropTypes.object,
}

export default TaskDefinitionSidebar
