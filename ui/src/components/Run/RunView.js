import React from "react"
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
import intentTypes from "../../constants/intentTypes"
import runStatusTypes from "../../constants/runStatusTypes"

const RunView = props => {
  return (
    <RunContext.Consumer>
      {({ data, runID }) => {
        let actions = [
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
          get(data, "status") !== runStatusTypes.stopped
        ) {
          actions.push({
            isLink: false,
            text: "Stop Run",
            intent: intentTypes.error,
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

export default RunView
