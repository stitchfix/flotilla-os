import * as React from "react"
import { withRouter, RouteComponentProps } from "react-router-dom"
import { get, Omit } from "lodash"
import BaseTaskForm, { TaskFormPayload } from "./BaseTaskForm"
import api from "../../api"
import {
  IFlotillaTaskDefinition,
  flotillaUIIntents,
  IFlotillaUIPopupProps,
  IFlotillaAPIError,
  flotillaUIRequestStates,
  IFlotillaEditTaskPayload,
} from "../../types"
import PopupContext from "../Popup/PopupContext"
import TaskContext from "../Task/TaskContext"
import { UpdateTaskYupSchema } from "./validation"
import Loader from "../styled/Loader"

export interface IProps {
  defaultValues: TaskFormPayload
  push: any
  renderPopup: (p: IFlotillaUIPopupProps) => void
  title: string
  definitionID: string
  requestData: () => void
}

export class UpdateTaskForm extends React.PureComponent<IProps> {
  preprocessValues = (values: TaskFormPayload): IFlotillaEditTaskPayload => ({
    command: values.command,
    env: values.env,
    group_name: values.group_name,
    image: values.image,
    memory: +values.memory,
    tags: values.tags,
  })

  handleSubmit = (values: TaskFormPayload): Promise<any> =>
    api.updateTask({
      definitionID: this.props.definitionID,
      values: this.preprocessValues(values),
    })

  handleSuccess = (res: IFlotillaTaskDefinition) => {
    const { push, definitionID, requestData } = this.props
    requestData()
    push(`/tasks/${definitionID}`)
  }

  handleFail = (error: IFlotillaAPIError) => {
    const { renderPopup } = this.props

    renderPopup({
      body: error.data,
      intent: flotillaUIIntents.ERROR,
      shouldAutohide: false,
      title: `An error occurred (Status Code: ${error.status})`,
    })
  }

  render() {
    const { defaultValues, title } = this.props
    return (
      <BaseTaskForm
        defaultValues={defaultValues}
        title={title}
        submitFn={this.handleSubmit}
        onSuccess={this.handleSuccess}
        onFail={this.handleFail}
        validateSchema={UpdateTaskYupSchema}
        shouldRenderAliasField={false}
      />
    )
  }
}

// This layer connects the UpdateTaskForm component to the TaskContext. We're
// doing this for easier testing - we don't really need to test the most
// outward component (the one connected to the Router and PopupContext below)
// but we need to ensure that we're sourcing the correct default values from
// the task context.
export const WithTaskContext: React.SFC<
  Pick<IProps, "push" | "renderPopup">
> = props => (
  <TaskContext.Consumer>
    {ctx => {
      if (ctx.requestState === flotillaUIRequestStates.READY) {
        return (
          <UpdateTaskForm
            push={props.push}
            renderPopup={props.renderPopup}
            defaultValues={{
              command: get(ctx, ["data", "command"], ""),
              env: get(ctx, ["data", "env"], []),
              group_name: get(ctx, ["data", "group_name"], ""),
              image: get(ctx, ["data", "image"], ""),
              memory: get(ctx, ["data", "memory"], 1024),
              tags: get(ctx, ["data", "tags"], []),
            }}
            title={`Update Task ${get(
              ctx,
              ["data", "alias"],
              ctx.definitionID
            )}`}
            definitionID={ctx.definitionID}
            requestData={ctx.requestData}
          />
        )
      }

      return <Loader />
    }}
  </TaskContext.Consumer>
)

// Connect to Router and PopupContext.
export default withRouter(
  (props: Omit<IProps, "renderPopup" | "push"> & RouteComponentProps<{}>) => (
    <PopupContext.Consumer>
      {popupCtx => (
        <WithTaskContext
          push={props.history.push}
          renderPopup={popupCtx.renderPopup}
        />
      )}
    </PopupContext.Consumer>
  )
)
