import * as React from "react"
import ReactResizeDetector from "react-resize-detector"
import { isEmpty, round } from "lodash"
import {
  NAVIGATION_HEIGHT_PX,
  DETAIL_VIEW_SIDEBAR_WIDTH_PX,
} from "../../helpers/styles"
import LogRenderer from "./LogRenderer"
import { IFlotillaUILogChunk } from "../../.."

interface IUnwrappedLogProcessorProps {
  logs: IFlotillaUILogChunk[]
}

interface ILogProcessorProps extends IUnwrappedLogProcessorProps {
  width: number
  height: number
}

/**
 * The intermediate component between LogRequester and LogRenderer. This
 * component is responsible for slicing the logs into smaller pieces, each of
 * which will be rendered into a LowRow component.
 */
class LogProcessor extends React.PureComponent<ILogProcessorProps> {
  static HACKY_CHAR_TO_PIXEL_RATIO = 37 / 300

  /**
   * Returns the max number of characters allowed per line.
   */
  getMaxLineLength = (): number =>
    round(this.props.width * LogProcessor.HACKY_CHAR_TO_PIXEL_RATIO)

  /**
   * Takes the `logs` prop (an array of LogChunk objects), splits each
   * LogChunk's log string according to the available width, and flattens it to
   * an array of strings, which it then passes to LogRenderer to render.
   */
  processLogs = (): string[] => {
    const { logs } = this.props

    if (isEmpty(logs)) return []

    const maxLineLength = this.getMaxLineLength()

    return logs.reduce(
      (acc: string[], chunk: IFlotillaUILogChunk): string[] => {
        // Split the chunk string by newline chars.
        const split = chunk.chunk.split("\n")

        // Loop through each split part of the chunk. For each part, if the
        // length of the string is greater than the maxLineLength variable, split
        // the part so each sub-part is less than maxLineLength. Otherwise, push
        // the part to the array to be returned.
        for (let i = 0; i < split.length; i++) {
          const str: string = split[i]

          if (str.length > maxLineLength) {
            for (let j = 0; j < str.length; j += maxLineLength) {
              acc.push(str.slice(j, j + maxLineLength))
            }
          } else {
            acc.push(str)
          }
        }

        return acc
      },
      []
    )
  }

  /**
   * Checks whether the dimensions have been set by ReactSizeDetector.
   */
  areDimensionsValid = (): boolean => {
    const { width, height } = this.props

    if (
      width === 0 ||
      width === undefined ||
      height === 0 ||
      height === undefined
    ) {
      return false
    }

    return true
  }

  render() {
    if (this.areDimensionsValid()) {
      // Only process logs if the dimensions are valid.
      const logs = this.processLogs()
      return <LogRenderer {...this.props} logs={logs} len={logs.length} />
    }

    return <span />
  }
}

const WrappedLogProcessor: React.SFC<IUnwrappedLogProcessorProps> = props => (
  <ReactResizeDetector
    handleHeight
    handleWidth
    refreshMode="throttle"
    refreshRate={500}
  >
    {(w: number, h: number) => {
      let height = h
      let width = w

      if (h === 0 || h === undefined) {
        height = window.innerHeight - NAVIGATION_HEIGHT_PX
      }

      if (w === 0 || w === undefined) {
        width = DETAIL_VIEW_SIDEBAR_WIDTH_PX
      }

      return <LogProcessor {...props} width={width} height={height} />
    }}
  </ReactResizeDetector>
)

export default WrappedLogProcessor
