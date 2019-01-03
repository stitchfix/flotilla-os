import * as React from "react"
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
import colors from "../../helpers/colors"

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
  background: ${colors.black[0]};
`

const iconProps = {
  style: { transform: `translateY(2px)` },
  size: 14,
}

interface IRunBarProps {
  onScrollToTopClick: () => void
  onScrollToBottomClick: () => void
  toggleShouldAutoscroll: () => void
  shouldAutoscroll: boolean
}

class RunBar extends React.PureComponent<IRunBarProps> {
  render() {
    return (
      <RunContext.Consumer>
        {({ data }) => (
          <RunBarContainer>
            <RunStatus
              onlyRenderIcon={false}
              status={get(data, "status")}
              exitCode={get(data, "exit_code")}
            />
            <ButtonGroup>
              {/* <Button onClick={this.props.onScrollToTopClick}>
                <ChevronsUp {...iconProps} />
              </Button>
              <Button onClick={this.props.onScrollToBottomClick}>
                <ChevronsDown {...iconProps} />
              </Button> */}
              <Button onClick={this.props.toggleShouldAutoscroll}>
                {this.props.shouldAutoscroll ? (
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
  }
}

export default RunBar
