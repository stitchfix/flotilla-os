import * as React from "react"
import { get, omit } from "lodash"
import KeyValues from "../styled/KeyValues"
import Tag from "../styled/Tag"
import { Pre } from "../styled/Monospace"
import TagGroup from "../styled/TagGroup"
import { IFlotillaTaskDefinition, IFlotillaEnv } from "../../.."

interface ITaskDefinitionSidebarProps {
  data: IFlotillaTaskDefinition | null
}

class TaskDefinitionSidebar extends React.PureComponent<
  ITaskDefinitionSidebarProps
> {
  static displayName = "TaskDefinitionSidebar"
  render() {
    const { data } = this.props
    return (
      <React.Fragment>
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
          items={
            data &&
            data.env &&
            data.env.reduce((acc: any, env: IFlotillaEnv): any => {
              acc[env.name] = <Pre>{env.value}</Pre>
              return acc
            }, {})
          }
        />
      </React.Fragment>
    )
  }
}

export default TaskDefinitionSidebar
