import * as React from "react"
import { isEmpty } from "lodash"
import { Plus } from "react-feather"
import Field from "../styled/Field"
import { Input } from "../styled/Inputs"
import NestedKeyValueRow from "../styled/NestedKeyValueRow"
import Button from "../styled/Button"

interface IKVFieldInputProps {
  addValue: any
  field: string
  isKeyRequired: boolean
  isValueRequired: boolean
  keyField: string
  valueField: string
}

interface IKVFieldInputState {
  keyValue: string
  valueValue: string
  isKeyInputFocused: boolean
  isValueInputFocused: boolean
}

class KVFieldInput extends React.PureComponent<
  IKVFieldInputProps,
  IKVFieldInputState
> {
  static displayName = "KVFieldInput"

  static defaultProps: Partial<IKVFieldInputProps> = {
    isKeyRequired: true,
    isValueRequired: false,
    keyField: "name",
    valueField: "value",
  }

  private keyInputRef = React.createRef<HTMLInputElement>()

  state = {
    keyValue: "",
    valueValue: "",
    isKeyInputFocused: false,
    isValueInputFocused: false,
  }

  componentDidMount() {
    window.addEventListener("keypress", this.handleKeypress)
  }

  /** Determines whether or not to add the user input to the value. */
  shouldAddField = (): boolean => {
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

  addField = (): void => {
    const { addValue, field, keyField, valueField } = this.props
    const { keyValue, valueValue } = this.state

    addValue(field, { [keyField]: keyValue, [valueField]: valueValue })
    this.resetState()
  }

  resetState = () => {
    this.setState({ keyValue: "", valueValue: "" }, () => {
      const keyInputNode = this.keyInputRef.current

      if (keyInputNode) {
        keyInputNode.focus()
      }
    })
  }

  handleKeypress = (evt: KeyboardEvent) => {
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
            ref={this.keyInputRef}
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

export default KVFieldInput
