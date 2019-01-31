import * as React from "react"
import Ansi from "ansi-to-react"
import { Pre } from "../styled/Monospace"
import Loader from "../styled/Loader"
import RunBar from "../Run/RunBar"
import { IFlotillaUILogChunk, flotillaUIIntents } from "../../.."
import { RUN_BAR_HEIGHT_PX } from "../../helpers/styles"

interface INonOptimizedLogRendererProps {
  logs: IFlotillaUILogChunk[]
  hasRunFinished: boolean
}

interface INonOptimizedLogRendererState {
  shouldAutoscroll: boolean
}

class NonOptimizedLogRenderer extends React.Component<
  INonOptimizedLogRendererProps,
  INonOptimizedLogRendererState
> {
  private CONTAINER_DIV = React.createRef<HTMLDivElement>()
  state = {
    shouldAutoscroll: true,
  }

  componentDidMount() {
    if (this.state.shouldAutoscroll) {
      this.scrollToBottom()
    }
  }

  componentDidUpdate(prevProps: INonOptimizedLogRendererProps) {
    if (
      this.state.shouldAutoscroll &&
      prevProps.logs.length !== this.props.logs.length
    ) {
      this.scrollToBottom()
    }
  }

  toggleShouldAutoscroll = (): void => {
    this.setState(prev => ({ shouldAutoscroll: !prev.shouldAutoscroll }))
  }

  scrollToTop = (): void => {
    const container = this.CONTAINER_DIV.current

    if (container) {
      container.scrollTop = 0
    }
  }

  scrollToBottom = (): void => {
    const container = this.CONTAINER_DIV.current

    if (container) {
      container.scrollTop = container.scrollHeight
    }
  }

  render() {
    const { logs, hasRunFinished } = this.props

    let loader = null

    if (!hasRunFinished) {
      loader = (
        <span
          style={{
            display: "flex",
            flexFlow: "row nowrap",
            justifyContent: "flex-start",
            alignItems: "center",
            width: "100%",
            padding: 24,
          }}
        >
          <Loader intent={flotillaUIIntents.PRIMARY} />
        </span>
      )
    }

    return (
      <div
        ref={this.CONTAINER_DIV}
        style={{ overflowY: "scroll", height: "100%" }}
      >
        <RunBar
          shouldAutoscroll={this.state.shouldAutoscroll}
          toggleShouldAutoscroll={this.toggleShouldAutoscroll}
          onScrollToTopClick={this.scrollToTop}
          onScrollToBottomClick={this.scrollToBottom}
        />
        <Pre style={{ paddingTop: RUN_BAR_HEIGHT_PX }}>
          {logs.map((l: IFlotillaUILogChunk) => (
            // Note: using a CSS classname here as the Ansi component does not
            // support a `style` prop. The style definition is located in
            // ui/src/index.html
            <Ansi key={l.lastSeen} className="flotilla-ansi">
              {l.chunk}
            </Ansi>
          ))}
          {loader}
        </Pre>
      </div>
    )
  }
}

export default NonOptimizedLogRenderer
