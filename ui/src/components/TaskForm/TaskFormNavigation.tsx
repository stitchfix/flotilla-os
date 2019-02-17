import * as React from "react"
import Navigation from "../Navigation/Navigation"
import { flotillaUIIntents, IFlotillaUIBreadcrumb } from "../../types"

interface IProps {
  isSubmitDisabled: boolean
  inFlight: boolean
  breadcrumbs?: IFlotillaUIBreadcrumb[]
}

const TaskFormNavigation: React.SFC<IProps> = props => {
  const { isSubmitDisabled, inFlight, breadcrumbs } = props
  return (
    <Navigation
      breadcrumbs={breadcrumbs}
      actions={[
        {
          isLink: false,
          text: "Cancel",
          buttonProps: {
            // onClick: goBack,
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

export default TaskFormNavigation
