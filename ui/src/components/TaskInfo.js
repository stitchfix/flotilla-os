import React from "react"
import { Card, FormGroup, Tag } from "aa-ui-components"
import { get, has } from "lodash"
import JsonView from "react-json-view"
import { reactJsonViewProps } from "../constants/"
import KeyValueContainer from "./KeyValueContainer"

const TaskInfo = ({ data }) => (
  <div className="flot-detail-view-sidebar">
    <KeyValueContainer header="Task Info">
      {({ json, collapsed }) => {
        if (json) {
          return <JsonView {...reactJsonViewProps} src={data} />
        }

        return (
          <div className="flot-detail-view-sidebar-card-content">
            <FormGroup isStatic label="Alias">
              {get(data, "alias", "...")}
            </FormGroup>
            <FormGroup isStatic label="Definition ID">
              {get(data, "definition_id", "...")}
            </FormGroup>
            <FormGroup isStatic label="Container Name">
              {get(data, "container_name", "...")}
            </FormGroup>
            <FormGroup isStatic label="Group Name">
              {get(data, "group_name", "...")}
            </FormGroup>
            <FormGroup isStatic label="Image">
              {get(data, "image", "...")}
            </FormGroup>
            <FormGroup isStatic label="Command">
              <pre style={{ fontSize: "0.9rem" }}>
                {get(data, "command", "...")}
              </pre>
            </FormGroup>
            <FormGroup isStatic label="Memory">
              {get(data, "memory", "...")}
            </FormGroup>
            <FormGroup isStatic label="Arn">
              {get(data, "arn", "...")}
            </FormGroup>
            <FormGroup isStatic label="Tags">
              <div className="flex ff-rw j-fs a-fs with-horizontal-child-margin">
                {get(data, "tags", [])
                  .filter(tag => tag !== "")
                  .map(tag => <Tag key={tag}>{tag}</Tag>)}
              </div>
            </FormGroup>
          </div>
        )
      }}
    </KeyValueContainer>
    <KeyValueContainer header="Environment Variables">
      {({ json, collapsed }) => {
        if (json) {
          return (
            <JsonView
              {...reactJsonViewProps}
              src={get(data, "env", []).reduce((acc, val) => {
                acc[val.name] = val.value
                return acc
              }, {})}
            />
          )
        }

        return (
          <div className="flot-detail-view-sidebar-card-content code">
            {get(data, "env", []).map((env, i) => (
              <FormGroup
                horizontal
                isStatic
                label={<Tag>{env.name}</Tag>}
                key={`env-${i}`}
              >
                <Tag>{env.value}</Tag>
              </FormGroup>
            ))}
          </div>
        )
      }}
    </KeyValueContainer>
  </div>
)

export default TaskInfo
