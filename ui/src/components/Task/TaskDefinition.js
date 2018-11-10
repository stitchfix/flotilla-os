import React, { Fragment } from "react"
import { Link } from "react-router-dom"
import JSONView from "react-json-view"
import { get, omit } from "lodash"
import TaskContext from "./TaskContext"
import * as requestStateTypes from "../../constants/requestStateTypes"
import View from "../styled/View"
import ViewHeader from "../styled/ViewHeader"
import Loader from "../styled/Loader"
import TaskHistoryTable from "./TaskHistoryTable"
import Button from "../styled/Button"
import intentTypes from "../../constants/intentTypes"
import DeleteTaskModal from "../Modal/DeleteTaskModal"
import ButtonGroup from "../styled/ButtonGroup"
import KeyValueContainer from "../styled/KeyValueContainer"
import FormGroup from "../styled/FormGroup"
import Tag from "../styled/Tag"
import reactJsonViewProps from "../../constants/reactJsonViewProps"
import ModalContext from "../Modal/ModalContext"

const TaskDefinitionSidebar = ({ data }) => {
  return (
    <Fragment>
      <KeyValueContainer header="Task Info">
        {({ json, collapsed }) => {
          if (json) {
            return <JSONView {...reactJsonViewProps} src={data} />
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
              <JSONView
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
                  isStatic
                  label={
                    <span className="code" style={{ color: "white" }}>
                      {env.name}
                    </span>
                  }
                  key={`env-${i}`}
                >
                  <span className="code" style={{ wordBreak: "break-all" }}>
                    {env.value}
                  </span>
                </FormGroup>
              ))}
            </div>
          )
        }}
      </KeyValueContainer>
    </Fragment>
  )
}

const TaskDefinition = props => {
  return (
    <TaskContext.Consumer>
      {ctx => {
        let title = <Loader mini />
        let actions
        let sidebar = <Loader />

        switch (ctx.requestState) {
          case requestStateTypes.READY:
            title = get(ctx, ["data", "alias"], "")
            actions = (
              <ButtonGroup>
                <Button
                  intent={intentTypes.error}
                  onClick={() => {
                    props.renderModal(
                      <DeleteTaskModal definitionID={ctx.definitionID} />
                    )
                  }}
                >
                  Delete
                </Button>
                <Link
                  to={`/tasks/${ctx.definitionID}/copy`}
                  className="pl-button"
                >
                  Copy
                </Link>
                <Link
                  to={`/tasks/${ctx.definitionID}/edit`}
                  className="pl-button"
                >
                  Edit
                </Link>
                <Link
                  to={`/tasks/${ctx.definitionID}/run`}
                  className="pl-button pl-intent-primary"
                >
                  Run
                </Link>
              </ButtonGroup>
            )
            sidebar = <TaskDefinitionSidebar data={ctx.data} />
            break
          case requestStateTypes.ERROR:
            title = "Error"
            sidebar = "blork"
            break
          case requestStateTypes.NOT_READY:
          default:
            title = "loading"
            sidebar = "loading"
            break
        }

        return (
          <View>
            <ViewHeader title={title} actions={actions} />
            <div>
              <div className="flot-detail-view-sidebar">{sidebar}</div>
              <TaskHistoryTable definitionID={ctx.definitionID} />
            </div>
          </View>
        )
      }}
    </TaskContext.Consumer>
  )
}

export default props => (
  <ModalContext.Consumer>
    {ctx => (
      <TaskDefinition
        {...omit(props, ["history", "location", "match", "staticContext"])}
        push={props.history.push}
        renderModal={ctx.renderModal}
        unrenderModal={ctx.unrenderModal}
      />
    )}
  </ModalContext.Consumer>
)
