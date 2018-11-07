import React, { Component } from "react"
import { isFunction } from "lodash"
import Select, { Creatable } from "react-select"
import asField from "./asField"
import FormGroup from "./FormGroup"

class ReduxFormGroupSelect extends Component {
  state = {
    touched: false,
  }
  // When creating a new option via React Select's `Creatable` component,
  // the internal Select's state will not update if the newly created
  // option is passed as a string value to the Select's `value` prop
  // (which, in every other scenario, it will always be passed as a
  // string). Thus, while the Select component's value prop is correct, it
  // will not display correctly. To solve this, when a new option is
  // created, the value should not be a string, but rather an object with
  // a `label` and `value` key (e.g. the typical format for Select's
  // `options` prop).
  isNewOption(value) {
    return !this.props.custom.options.map(opt => opt.value).includes(value)
  }
  getValue() {
    const { input, custom } = this.props

    if (input.value === "") {
      return null
    }

    if (!!custom.multi && Array.isArray(input.value)) {
      return input.value.filter(val => !!val).map(val => {
        if (this.isNewOption(val)) {
          return { label: val, value: val }
        }
        return val
      })
    }

    if (this.isNewOption(input.value)) {
      return { label: input.value, value: input.value }
    }

    return input.value
  }
  renderInput() {
    const { touched } = this.state
    const { input, custom } = this.props

    const props = {
      ...custom,
      value: this.getValue(),
      options: custom.options,
      multi: custom.multi,
      placeholder: custom.placeholder,
      clearable: custom.clearable,
      onChange: option => {
        if (!!option) {
          // Differentiate between single/multi values
          if (Array.isArray(option)) {
            input.onChange(option.map(o => o.value))
          } else {
            input.onChange(option.value)
          }
        } else {
          input.onChange(null)
        }
      },
      onFocus: evt => {
        if (!touched) {
          this.setState({ touched: true })
        }

        if (isFunction(custom.onFocus)) {
          custom.onFocus(evt)
        }
      },
      onBlur: evt => {
        if (!touched) {
          this.setState({ touched: true })
        }

        if (isFunction(custom.onBlur)) {
          custom.onBlur(evt)
        }
      },
    }

    if (custom.allowCreate) {
      return <Creatable {...props} />
    }

    return <Select {...props} />
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

export default asField(ReduxFormGroupSelect)
