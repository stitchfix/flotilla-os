import * as React from "react"
import { get } from "lodash"
import { Tag, Colors, Checkbox, Intent } from "@blueprintjs/core"
import { Task, UpdateTaskPayload } from "../types"
import api from "../api"
import Toaster from "./Toaster"
import Request, { ChildProps } from "./Request"

type Props = {
  task: Task
} & ChildProps<Task, { definitionID: string; data: UpdateTaskPayload }>

class ARASwitch extends React.Component<Props> {
  constructor(props: Props) {
    super(props)
    this.handleChange = this.handleChange.bind(this)
  }

  handleChange() {
    const { task, request } = this.props

    let enabled: boolean
    if (this.isEnabled()) {
      enabled = false
    } else {
      enabled = true
    }

    request({
      definitionID: task.definition_id,
      data: {
        env: task.env,
        image: task.image,
        group_name: task.group_name,
        memory: task.memory,
        cpu: task.cpu,
        command: task.command,
        tags: task.tags,
        adaptive_resource_allocation: enabled,
      },
    })
  }

  isEnabled() {
    return get(this.props, "adaptive_resource_allocation", false) === true
  }

  render() {
    const enabled = this.isEnabled()
    return (
      <div style={{ display: "flex", alignItems: "center" }}>
        <Checkbox
          checked={enabled}
          onChange={this.handleChange}
          style={{ marginBottom: 0 }}
        />
        <Tag
          style={{
            background: enabled ? Colors.ROSE5 : "",
            color: enabled ? Colors.WHITE : "",
            cursor: "default",
          }}
        >
          {enabled ? "Enabled" : "Disabled"}
        </Tag>
      </div>
    )
  }
}

type ConnectedProps = {
  task: Task
  request: (opts: { definitionID: string }) => void
}

const Connected: React.FC<ConnectedProps> = ({ task, request }) => (
  <Request<Task, { definitionID: string; data: UpdateTaskPayload }>
    requestFn={api.updateTask}
    shouldRequestOnMount={false}
    onSuccess={(data: Task) => {
      Toaster.show({
        message: `ARA enabled for ${data.alias}!`,
        intent: Intent.SUCCESS,
      })
      // Re-request data.
      request({ definitionID: data.definition_id })
    }}
    onFailure={() => {
      Toaster.show({
        message: "An error occurred.",
        intent: Intent.DANGER,
      })
    }}
  >
    {requestProps => <ARASwitch task={task} {...requestProps} />}
  </Request>
)

export default Connected
