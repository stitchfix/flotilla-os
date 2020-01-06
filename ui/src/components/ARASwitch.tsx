import * as React from "react"
import { get } from "lodash"
import { Switch, Tag, Colors, Tooltip, Checkbox } from "@blueprintjs/core"
import { Task } from "../types"
import api from "../api"

type Props = Task

class ARASwitch extends React.Component<Props> {
  constructor(props: Props) {
    super(props)
    this.handleChange = this.handleChange.bind(this)
  }
  handleChange() {
    let enabled: boolean
    if (this.isEnabled()) {
      enabled = false
    } else {
      enabled = true
    }

    api
      .updateTask({
        definitionID: this.props.definition_id,
        data: {
          env: this.props.env,
          image: this.props.image,
          group_name: this.props.group_name,
          memory: this.props.memory,
          cpu: this.props.cpu,
          command: this.props.command,
          tags: this.props.tags,
          adaptiveResourceAllocation: enabled,
        },
      })
      .then(res => {})
  }

  isEnabled() {
    return get(this.props, "adaptiveResourceAllocation", false) === true
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

export default ARASwitch
