import React, { Component } from "react"
import FormGroup from "./FormGroup"
import asField from "./asField"

class ReduxFormGroupInput extends Component {
  state = {
    touched: false,
  }
  renderInput() {
    const { input, custom } = this.props

    return (
      <input
        disabled={custom.disabled}
        className="pl-input"
        type={custom.type}
        value={input.value}
        placeholder={custom.placeholder}
        onChange={evt => {
          input.onChange(evt.target.value)
        }}
        onBlur={() => {
          this.setState({ touched: true })
        }}
      />
    )
  }
  render() {
    const { touched } = this.state
    const {
      meta: { error },
      custom: { label, isRequired, style, description, ...rest },
    } = this.props

    const hasError = !!(touched && !!error)
    const errorText = !!touched && !!error ? error : ""

    return (
      <FormGroup
        label={label}
        input={this.renderInput()}
        isRequired={isRequired}
        hasError={hasError}
        errorText={errorText}
        style={style}
        description={description}
        {...rest}
      />
    )
  }
}

export default asField(ReduxFormGroupInput)
