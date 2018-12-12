import * as React from "react"
import {
  FieldContainer,
  FieldLabel,
  FieldDescription,
  FieldError,
} from "../styled/Field"

interface IKVFieldContainerProps {
  description?: string
  error?: any
  isRequired: boolean
  label?: string
}

class KVFieldContainer extends React.PureComponent<IKVFieldContainerProps> {
  static defaultProps: Partial<IKVFieldContainerProps> = {
    isRequired: false,
  }

  render() {
    const { children, description, error, isRequired, label } = this.props
    return (
      <FieldContainer>
        {!!label && <FieldLabel isRequired={isRequired}>{label}</FieldLabel>}
        {!!description && (
          <span style={{ marginBottom: 8, marginTop: 0 }}>
            <FieldDescription>{description}</FieldDescription>
          </span>
        )}
        {!!error && <FieldError>{error}</FieldError>}
        {children}
      </FieldContainer>
    )
  }
}

export default KVFieldContainer
