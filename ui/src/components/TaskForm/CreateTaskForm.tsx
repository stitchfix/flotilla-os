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
} from "../../types"
import PopupContext from "../Popup/PopupContext"
import { CreateTaskYupSchema } from "./validation"

interface IProps {
  defaultValues: TaskFormPayload
  title: string
  push: any
  renderPopup: (p: IFlotillaUIPopupProps) => void
}

export class CreateTaskForm extends React.PureComponent<IProps> {
  static defaultProps: Partial<IProps> = {
    defaultValues: {
      alias: "",
      group_name: "",
      image: "",
      memory: 1024,
      command: "",
    },
    title: "Create New Task",
  }

  preprocessValues = (values: TaskFormPayload) => ({
    alias: get(values, "alias", ""),
    command: values.command,
    env: values.env,
    group_name: values.group_name,
    image: values.image,
    memory: +values.memory,
    tags: values.tags,
  })

  handleSubmit = (values: TaskFormPayload): Promise<any> =>
    api.createTask({ values: this.preprocessValues(values) })

  handleSuccess = (res: IFlotillaTaskDefinition) => {
    this.props.push(`/tasks/${res.definition_id}`)
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
        validateSchema={CreateTaskYupSchema}
      />
    )
  }
}

export default withRouter(
  (props: Omit<IProps, "renderPopup" | "push"> & RouteComponentProps<{}>) => (
    <PopupContext.Consumer>
      {ctx => (
        <CreateTaskForm
          push={props.history.push}
          renderPopup={ctx.renderPopup}
          defaultValues={props.defaultValues}
          title={props.title}
        />
      )}
    </PopupContext.Consumer>
  )
)
