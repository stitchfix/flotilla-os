import * as React from "react"
import { Tag, Tooltip, Icon, Intent } from "@blueprintjs/core"
import CopyToClipboard from "react-copy-to-clipboard"

type Props = { rawValue: string }
type State = { isCopied: boolean }

class CopyableAttributeValue extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.handleCopy = this.handleCopy.bind(this)
  }

  state = {
    isCopied: false,
  }

  handleCopy() {
    this.setState({ isCopied: true })
  }

  render() {
    return (
      <Tooltip
        content={
          <div>
            Click to copy to clipboard
            {this.state.isCopied && (
              <Icon
                icon="confirm"
                intent={Intent.SUCCESS}
                style={{ marginLeft: 6 }}
              />
            )}
          </div>
        }
      >
        <CopyToClipboard text={this.props.rawValue} onCopy={this.handleCopy}>
          <div style={{ cursor: "pointer" }}>{this.props.children}</div>
        </CopyToClipboard>
      </Tooltip>
    )
  }
}

const Attribute: React.FunctionComponent<{
  name: React.ReactNode
  value: React.ReactNode
  containerStyle?: object
  isCopyable?: boolean
  rawValue?: string
  description?: React.ReactElement
  isNew?: boolean
}> = ({
  name,
  value,
  containerStyle,
  isCopyable,
  rawValue,
  description,
  isNew,
}) => (
  <div
    className="flotilla-attribute-container"
    style={containerStyle ? containerStyle : {}}
  >
    <div className="flotilla-attribute-name">
      <div>{name}</div>
      {description && (
        <Tooltip content={description}>
          <Icon icon="info-sign" iconSize={14} />
        </Tooltip>
      )}
      {isNew && <Tag intent={Intent.DANGER}>New!</Tag>}
    </div>
    {isCopyable && rawValue ? (
      <CopyableAttributeValue rawValue={rawValue}>
        <div className="flotilla-attribute-value">{value}</div>
      </CopyableAttributeValue>
    ) : (
      <div className="flotilla-attribute-value">{value}</div>
    )}
  </div>
)

export default Attribute
