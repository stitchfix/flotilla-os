import React from "react"
import styled from "styled-components"
import { get } from "lodash"
import Navigation from "../Navigation/Navigation"
import RunContext from "./RunContext"
import RunSidebar from "./RunSidebar"
import LogRequester from "./LogRequester"
import View from "../styled/View"
import intentTypes from "../../constants/intentTypes"

const RunViewContainer = styled.div`
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: stretch;
  width: 100%;
`

const RunView = props => {
  return (
    <RunContext.Consumer>
      {ctx => {
        const actions = [
          {
            isLink: true,
            href: "/",
            text: "Retry",
          },
          {
            isLink: false,
            text: "Stop Run",
            intent: intentTypes.error,
          },
        ]

        const breadcrumbs = [
          {
            text: get(ctx, ["data", "alias"], ""),
            href: `/tasks/${get(ctx, ["data", "definition_id"])}`,
          },
          {
            text: ctx.runID,
            href: `/runs/${ctx.runID}`,
          },
        ]

        return (
          <View>
            <Navigation actions={actions} breadcrumbs={breadcrumbs} />
            <RunViewContainer>
              <RunSidebar />
              <div style={{ flex: 1 }}>
                <LogRequester
                  runID={ctx.runID}
                  status={get(ctx, ["data", "status"])}
                />
              </div>
            </RunViewContainer>
          </View>
        )
      }}
    </RunContext.Consumer>
  )
}

export default RunView
