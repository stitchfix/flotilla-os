import React, { Component } from 'react'
import { FormGroup, ReactSelectWrapper } from './'

export default class ReduxFormGroup extends Component {
  state = {
    touched: false
  }
  render() {
    const {
      // Props from Redux Form
      input: { value, onChange },
      meta: { error },
      // Custom Props
      custom
    } = this.props
    const { touched } = this.state

    const {
      label,
      isRequired,
      inputType,
      selectOpts,
      multi,
      style,
    } = custom

    let input = <span />

    switch (inputType) {
      case 'select':
        input = (
          <ReactSelectWrapper
            value={value}
            onChange={(o) => { onChange(o) }}
            options={selectOpts}
            multi={multi}
            onFocus={() => {
              if (!touched) { this.setState({ touched: true }) }
            }}
            onBlur={() => {
              if (!touched) { this.setState({ touched: true }) }
            }}
            {...custom}
          />
        )
        break
      case 'textarea':
        input = (
          <textarea
            className="textarea code"
            value={value}
            onChange={(evt) => { onChange(evt.target.value) }}
            onBlur={() => { this.setState({ touched: true }) }}
            {...custom}
          />
        )
        break
      case 'input':
      default:
        input = (
          <input
            className="input"
            value={value}
            onChange={(evt) => { onChange(evt.target.value) }}
            onBlur={() => { this.setState({ touched: true }) }}
            {...custom}
          />
        )
        break
    }

    const hasError = !!(touched && !!error)
    const errorText = (!!touched && !!error) ? error : ''

    return (
      <FormGroup
        label={label}
        input={input}
        isRequired={isRequired}
        hasError={hasError}
        errorText={errorText}
        style={style}
      />
    )
  }
}
