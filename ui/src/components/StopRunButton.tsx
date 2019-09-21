import * as React from "react"
import { Button, Dialog, Intent, Classes } from "@blueprintjs/core"
import Request, { ChildProps } from "./Request"
import api from "../api"
import Toaster from "./Toaster"
import { withRouter, RouteComponentProps } from "react-router-dom"
import ErrorCallout from "./ErrorCallout"

type Args = { definitionID: string; runID: string }
export type Props = ChildProps<any, Args> & ConnectedProps
type State = { isOpen: boolean }

export class StopRunButton extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.handleSubmitClick = this.handleSubmitClick.bind(this)
    this.openDialog = this.openDialog.bind(this)
    this.closeDialog = this.closeDialog.bind(this)
  }

  state = {
    isOpen: false,
  }

  openDialog() {
    this.setState({ isOpen: true })
  }

  closeDialog() {
    this.setState({ isOpen: false })
  }

  handleSubmitClick() {
    this.props.request({
      definitionID: this.props.definitionID,
      runID: this.props.runID,
    })
  }

  render() {
    const { error, isLoading } = this.props
    return (
      <>
        <Button intent={Intent.DANGER} onClick={this.openDialog}>
          Stop
        </Button>
        <Dialog isOpen={this.state.isOpen}>
          <div className={Classes.DIALOG_BODY}>
            {error && <ErrorCallout error={error} />}
            <span>Are you sure you want to stop this run?</span>
          </div>
          <div className={Classes.DIALOG_FOOTER}>
            <div className={Classes.DIALOG_FOOTER_ACTIONS}>
              <Button onClick={this.closeDialog}>Close</Button>
              <Button
                loading={isLoading}
                intent={Intent.DANGER}
                onClick={this.handleSubmitClick}
                id="flotillaStopRunSubmitButton"
              >
                Stop
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
  runID: string
}

const Connected: React.FunctionComponent<
  RouteComponentProps & ConnectedProps
> = ({ runID, definitionID, history }) => (
  <Request<any, Args>
    requestFn={api.stopRun}
    initialRequestArgs={{ runID, definitionID }}
    shouldRequestOnMount={false}
    onSuccess={() => {
      Toaster.show({
        message: "Run stopped!",
        intent: Intent.SUCCESS,
      })
      history.push(`/tasks/${definitionID}`)
    }}
    onFailure={() => {
      Toaster.show({
        message: "An error occurred.",
        intent: Intent.DANGER,
      })
    }}
  >
    {requestProps => (
      <StopRunButton
        {...requestProps}
        runID={runID}
        definitionID={definitionID}
      />
    )}
  </Request>
)

export default withRouter(Connected)
