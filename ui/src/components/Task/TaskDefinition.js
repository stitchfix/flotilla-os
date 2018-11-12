import React from "react"
import { get, omit } from "lodash"
import TaskContext from "./TaskContext"
import * as requestStateTypes from "../../constants/requestStateTypes"
import View from "../styled/View"
import ViewHeader from "../styled/ViewHeader"
import Loader from "../styled/Loader"
import TaskHistoryTable from "./TaskHistoryTable"
import Button from "../styled/Button"
import ButtonLink from "../styled/ButtonLink"
import intentTypes from "../../constants/intentTypes"
import DeleteTaskModal from "../Modal/DeleteTaskModal"
import ButtonGroup from "../styled/ButtonGroup"
import ModalContext from "../Modal/ModalContext"
import { TaskDefinitionView } from "../styled/TaskDefinition"
import TaskDefinitionSidebar from "./TaskDefinitionSidebar"

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
                <ButtonLink to={`/tasks/${ctx.definitionID}/copy`}>
                  Copy
                </ButtonLink>
                <ButtonLink to={`/tasks/${ctx.definitionID}/edit`}>
                  Edit
                </ButtonLink>
                <ButtonLink to={`/tasks/${ctx.definitionID}/run`}>
                  Run
                </ButtonLink>
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
            <TaskDefinitionView>
              <div>{sidebar}</div>
              <TaskHistoryTable definitionID={ctx.definitionID} />
            </TaskDefinitionView>
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
