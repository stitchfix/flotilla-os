import * as React from "react"
import { withRouter, RouteComponentProps } from "react-router-dom"
import Navigation from "../Navigation/Navigation"
import { flotillaUIIntents, IFlotillaUIBreadcrumb } from "../../types"
import { Omit } from "lodash"

export interface IProps {
  isSubmitDisabled: boolean
  inFlight: boolean
  breadcrumbs?: IFlotillaUIBreadcrumb[]
  goBack: () => void
}

export const TaskFormNavigation: React.SFC<IProps> = props => {
  const { isSubmitDisabled, inFlight, breadcrumbs, goBack } = props
  return (
    <Navigation
      breadcrumbs={breadcrumbs}
      actions={[
        {
          isLink: false,
          text: "Cancel",
          buttonProps: {
            onClick: goBack,
          },
        },
        {
          isLink: false,
          text: "Submit",
          buttonProps: {
            type: "submit",
            intent: flotillaUIIntents.PRIMARY,
            isDisabled: isSubmitDisabled,
            isLoading: !!inFlight,
          },
        },
      ]}
    />
  )
}

const ConnectedTaskFormNavigation: React.ComponentType<any> = withRouter(
  (props: Omit<IProps, "goBack"> & RouteComponentProps<{}>) => (
    <TaskFormNavigation
      isSubmitDisabled={props.isSubmitDisabled}
      inFlight={props.inFlight}
      breadcrumbs={props.breadcrumbs}
      goBack={props.history.goBack}
    />
  )
)

ConnectedTaskFormNavigation.displayName = "ConnectedTaskFormNavigation"

export default ConnectedTaskFormNavigation
