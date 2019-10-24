import * as React from "react"
import { Button, Dialog, Intent, Classes } from "@blueprintjs/core"
import { withRouter, RouteComponentProps } from "react-router-dom"
import Request, { ChildProps } from "./Request"
import api from "../api"
import Toaster from "./Toaster"
import ErrorCallout from "./ErrorCallout"

type Args = { definitionID: string }
export type Props = ChildProps<any, Args> & ConnectedProps
type State = { isOpen: boolean }

export class DeleteTaskButton extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.handleSubmitClick = this.handleSubmitClick.bind(this)
    this.openDialog = this.openDialog.bind(this)
    this.closeDialog = this.closeDialog.bind(this)
  }
  state = {
    isOpen: false,
  }

  handleSubmitClick() {
    this.props.request({ definitionID: this.props.definitionID })
  }

  openDialog() {
    this.setState({ isOpen: true })
  }

  closeDialog() {
    this.setState({ isOpen: false })
  }

  render() {
    const { isLoading, error } = this.props

    return (
      <>
        <Button
          intent={Intent.DANGER}
          onClick={this.openDialog}
          rightIcon="trash"
        >
          Delete
        </Button>
        <Dialog isOpen={this.state.isOpen}>
          <div className={Classes.DIALOG_BODY}>
            {error && <ErrorCallout error={error} />}
            <span>Are you sure you want to delete this task?</span>
          </div>
          <div className={Classes.DIALOG_FOOTER}>
            <div className={Classes.DIALOG_FOOTER_ACTIONS}>
              <Button onClick={this.closeDialog}>Close</Button>
              <Button
                loading={isLoading}
                intent={Intent.DANGER}
                onClick={this.handleSubmitClick}
                id="flotillaDeleteTaskSubmitButton"
              >
                Delete
              </Button>
            </div>
          </div>
        </Dialog>
      </>
    )
  }
}

type ConnectedProps = {
  definitionID: string
}

const Connected: React.FunctionComponent<
  RouteComponentProps & ConnectedProps
> = ({ definitionID, history }) => (
  <Request<any, Args>
    requestFn={api.deleteTask}
    initialRequestArgs={{ definitionID }}
    shouldRequestOnMount={false}
    onSuccess={() => {
      Toaster.show({
        message: "Task deleted!",
        intent: Intent.SUCCESS,
      })
      history.push(`/tasks`)
    }}
    onFailure={() => {
      Toaster.show({
        message: "An error occurred.",
        intent: Intent.DANGER,
      })
    }}
  >
    {requestProps => (
      <DeleteTaskButton {...requestProps} definitionID={definitionID} />
    )}
  </Request>
)

export default withRouter(Connected)
