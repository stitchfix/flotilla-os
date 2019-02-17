import * as React from "react"
import { get, has } from "lodash"
import Navigation from "../Navigation/Navigation"
import RunContext from "./RunContext"
import RunSidebar from "./RunSidebar"
import LogRequester from "../Log/LogRequester"
import View from "../styled/View"
import {
  DetailViewContainer,
  DetailViewContent,
  DetailViewSidebar,
} from "../styled/DetailView"
import StopRunModal from "../Modal/StopRunModal"
import ModalContext from "../Modal/ModalContext"
import {
  flotillaRunStatuses,
  flotillaUIIntents,
  IFlotillaUINavigationLink,
} from "../../types"

interface IUnwrappedRunViewProps {
  renderModal: (m: React.ReactNode) => void
}

class UnwrappedRunView extends React.PureComponent<IUnwrappedRunViewProps> {
  render() {
    const { renderModal } = this.props

    return (
      <RunContext.Consumer>
        {({ data, runID }) => {
          let actions: IFlotillaUINavigationLink[] = [
            {
              isLink: true,
              href: {
                pathname: `/tasks/${get(data, "definition_id", "")}/run`,
                state: {
                  env: get(data, "env"),
                  cluster: get(data, "cluster"),
                },
              },
              text: "Retry",
            },
          ]

          if (
            has(data, "status") &&
            get(data, "status") !== flotillaRunStatuses.STOPPED
          ) {
            actions.push({
              isLink: false,
              text: "Stop Run",
              buttonProps: {
                intent: flotillaUIIntents.ERROR,
                onClick: (evt: React.SyntheticEvent) => {
                  renderModal(
                    <StopRunModal
                      defintionID={get(data, "defintionID", "")}
                      runID={runID}
                    />
                  )
                },
              },
            })
          }

          const breadcrumbs = [
            {
              text: get(data, "alias", ""),
              href: `/tasks/${get(data, "definition_id", "")}`,
            },
            {
              text: runID,
              href: `/runs/${runID}`,
            },
          ]

          return (
            <View>
              <Navigation actions={actions} breadcrumbs={breadcrumbs} />
              <DetailViewContainer>
                <DetailViewContent>
                  <LogRequester runID={runID} status={get(data, "status")} />
                </DetailViewContent>
                <DetailViewSidebar>
                  <RunSidebar />
                </DetailViewSidebar>
              </DetailViewContainer>
            </View>
          )
        }}
      </RunContext.Consumer>
    )
  }
}

class WrappedRunView extends React.PureComponent<{}> {
  render() {
    return (
      <ModalContext.Consumer>
        {ctx => <UnwrappedRunView renderModal={ctx.renderModal} />}
      </ModalContext.Consumer>
    )
  }
}

export default WrappedRunView
