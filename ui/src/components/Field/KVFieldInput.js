import React, { Component } from "react"
import PropTypes from "prop-types"
import { isEmpty } from "lodash"
import { Plus } from "react-feather"
import Field from "../styled/Field"
import { Input } from "../styled/Inputs"
import NestedKeyValueRow from "../styled/NestedKeyValueRow"
import Button from "../styled/Button"
import intentTypes from "../../helpers/intentTypes"

class KVFieldInput extends Component {
  state = {
    keyValue: "",
    valueValue: "",
    isKeyInputFocused: false,
    isValueInputFocused: false,
  }

  componentDidMount() {
    window.addEventListener("keypress", this.handleKeypress)
  }

  shouldAddField = () => {
    const { isKeyRequired, isValueRequired } = this.props
    const {
      keyValue,
      valueValue,
      isKeyInputFocused,
      isValueInputFocused,
    } = this.state

    if (!isKeyInputFocused && !isValueInputFocused) {
      return false
    }

    if (isKeyRequired === true && isEmpty(keyValue)) {
      return false
    }

    if (isValueRequired === true && isEmpty(valueValue)) {
      return false
    }

    return true
  }

  addField = () => {
    const { addValue, field, keyField, valueField } = this.props
    const { keyValue, valueValue } = this.state

    addValue(field, { [keyField]: keyValue, [valueField]: valueValue })
    this.resetState()
  }

  resetState = () => {
    this.setState({ keyValue: "", valueValue: "" }, () => {
      this.keyInput.focus()
    })
  }

  handleKeypress = evt => {
    const { isKeyInputFocused, isValueInputFocused } = this.state

    if (isKeyInputFocused || isValueInputFocused) {
      if (evt.keyCode === 13) {
        evt.preventDefault()
        evt.stopPropagation()

        if (this.shouldAddField()) {
          this.addField()
        }
      }
    }
  }

  toggleKeyInputFocus = () => {
    this.setState(prevState => ({
      isKeyInputFocused: !prevState.isKeyInputFocused,
    }))
  }

  toggleValueInputFocus = () => {
    this.setState(prevState => ({
      isValueInputFocused: !prevState.isValueInputFocused,
    }))
  }

  render() {
    const { isKeyRequired, isValueRequired } = this.props

    return (
      <NestedKeyValueRow>
        <Field
          label="Key"
          isRequired={isKeyRequired}
          description="Press enter to add."
        >
          <Input
            type="text"
            value={this.state.keyValue}
            onChange={evt => {
              this.setState({ keyValue: evt.target.value })
            }}
            ref={x => {
              this.keyInput = x
            }}
            onFocus={this.toggleKeyInputFocus}
            onBlur={this.toggleKeyInputFocus}
          />
        </Field>
        <Field label="Value" isRequired={isValueRequired}>
          <Input
            type="text"
            value={this.state.valueValue}
            onChange={evt => {
              this.setState({ valueValue: evt.target.value })
            }}
            onFocus={this.toggleValueInputFocus}
            onBlur={this.toggleValueInputFocus}
          />
        </Field>
        <div style={{ transform: "translateY(24px)" }}>
          <Button onClick={this.addField} type="button">
            <Plus size={14} />
          </Button>
        </div>
      </NestedKeyValueRow>
    )
  }
}

KVFieldInput.displayName = "KVFieldInput"
KVFieldInput.propTypes = {
  addValue: PropTypes.func.isRequired,
  field: PropTypes.string.isRequired,
  isKeyRequired: PropTypes.bool.isRequired,
  isValueRequired: PropTypes.bool.isRequired,
  keyField: PropTypes.string.isRequired,
  valueField: PropTypes.string.isRequired,
}
KVFieldInput.defaultProps = {
  isKeyRequired: true,
  isValueRequired: false,
  keyField: "name",
  valueField: "value",
}

export default KVFieldInput
