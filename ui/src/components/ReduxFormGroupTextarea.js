import React, { Component } from "react"
import FormGroup from "./FormGroup"
import asField from "./asField"

class ReduxFormGroupTextarea extends Component {
  state = {
    touched: false,
  }

  renderInput() {
    const { input: { value, onChange }, custom: { disabled } } = this.props

    return (
      <textarea
        disabled={disabled}
        className="pl-textarea"
        value={value}
        onChange={evt => {
          onChange(evt.target.value)
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

export default asField(ReduxFormGroupTextarea)
