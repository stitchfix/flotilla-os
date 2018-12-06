import React from "react"
import { CheckSquare, Square, ChevronsUp, ChevronsDown } from "react-feather"
import styled from "styled-components"
import { get } from "lodash"
import RunContext from "./RunContext"
import RunStatus from "./RunStatus"
import Button from "../styled/Button"
import ButtonGroup from "../styled/ButtonGroup"
import {
  NAVIGATION_HEIGHT_PX,
  DEFAULT_BORDER,
  DETAIL_VIEW_SIDEBAR_WIDTH_PX,
  SPACING_PX,
  RUN_BAR_HEIGHT_PX,
} from "../../helpers/styles"

const RunBarContainer = styled.div`
  height: ${RUN_BAR_HEIGHT_PX}px;
  display: flex;
  flex-flow: row nowrap;
  justify-content: space-between;
  align-items: center;
  border-bottom: ${DEFAULT_BORDER};
  padding: 0 ${SPACING_PX * 2}px;
  position: fixed;
  top: ${NAVIGATION_HEIGHT_PX}px;
  left: ${DETAIL_VIEW_SIDEBAR_WIDTH_PX}px;
  right: 0;
`

const iconProps = {
  style: { transform: `translateY(2px)` },
  size: 14,
}

const RunBar = props => (
  <RunContext.Consumer>
    {({ data }) => (
      <RunBarContainer>
        <RunStatus
          onlyRenderIcon={false}
          status={get(data, "status")}
          exitCode={get(data, "exit_code")}
        />
        <ButtonGroup>
          <Button onClick={props.onScrollToTopClick}>
            <ChevronsUp {...iconProps} />
          </Button>
          <Button onClick={props.onScrollToBottomClick}>
            <ChevronsDown {...iconProps} />
          </Button>
          <Button onClick={props.toggleShouldAutoscroll}>
            {props.shouldAutoscroll ? (
              <CheckSquare {...iconProps} />
            ) : (
              <Square {...iconProps} />
            )}
            <span style={{ marginLeft: 4 }}>Autoscroll</span>
          </Button>
        </ButtonGroup>
      </RunBarContainer>
    )}
  </RunContext.Consumer>
)

export default RunBar
