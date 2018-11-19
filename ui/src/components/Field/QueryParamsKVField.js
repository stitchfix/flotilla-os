import React, { Component } from "react"
import PropTypes from "prop-types"
import { X } from "react-feather"
import { get, pick, isArray } from "lodash"
import { FieldText } from "./FieldText"
import Button from "../styled/Button"
import Field from "../styled/Field"
import NestedKeyValueRow from "../styled/NestedKeyValueRow"
import intentTypes from "../../constants/intentTypes"
import KVFieldInput from "./KVFieldInput"
import QueryParams from "../QueryParams/QueryParams"
import {
  SHARED_KV_FIELD_PROPS,
  SHARED_KV_FIELD_DEFAULT_PROPS,
} from "../../utils/kvFieldHelpers"

class UnwrappedQueryParamsKVField extends Component {
  /** Handles input events for key fields. */
  handleKeyChange = (value, index) => {
    const { keyField } = this.props

    this.handleChange({
      key: keyField,
      value,
      index,
    })
  }

  /** Handles input events for value fields. */
  handleValueChange = (value, index) => {
    const { valueField } = this.props

    this.handleChange({
      key: valueField,
      value,
      index,
    })
  }

  /**
   * Injects a new value into the values array prop then calls this.setValues.
   */
  handleChange = ({ key, value, index }) => {
    const values = this.getValues()
    const next = [
      ...values.slice(0, index),
      {
        ...values[index],
        [key]: value,
      },
      ...values.slice(index + 1),
    ]

    this.setValues(next)
  }

  /** Appends the newly added KV pair to the query[field] array. */
  handleAddField = (_, kv) => {
    const values = this.getValues()
    this.setValues([...values, kv])
  }

  /** Removes a value specified by index. */
  handleRemoveClick = index => {
    const values = this.getValues()
    this.setValues([...values.slice(0, index), ...values.slice(index + 1)])
  }

  /**
   * Stringifies a key value object (e.g. { name: "", value: ""}) into a string
   * delimited by the keyValueDelimiterChar prop (e.g. "key|value").
   */
  stringifyValue = obj => {
    const { keyField, valueField, keyValueDelimiterChar } = this.props
    return `${obj[keyField]}${keyValueDelimiterChar}${obj[valueField]}`
  }

  /**
   * Parses a key value string object and transforms it into an object.
   */
  parseValue = str => {
    const { keyField, valueField, keyValueDelimiterChar } = this.props
    const split = str.split(keyValueDelimiterChar)
    return {
      [keyField]: split[0],
      [valueField]: split[1],
    }
  }

  /** Calls props.setQueryParams to set new values. */
  setValues = values => {
    const { setQueryParams, field } = this.props

    setQueryParams({ [field]: values.map(this.stringifyValue) })
  }

  /** Transforms each value in the values prop to an object. */
  getValues = () => {
    const { values } = this.props

    return values.map(this.parseValue)
  }

  render() {
    const {
      label,
      keyField,
      isKeyRequired,
      isValueRequired,
      valueField,
    } = this.props

    return (
      <Field label={label}>
        {this.getValues().map((v, i) => {
          return (
            <NestedKeyValueRow>
              <FieldText
                field={keyField}
                isRequired={isKeyRequired}
                onChange={value => {
                  this.handleKeyChange(value, i)
                }}
                value={get(v, keyField, "")}
                shouldDebounce
              />
              <FieldText
                field={valueField}
                isRequired={isValueRequired}
                onChange={value => {
                  this.handleValueChange(value, i)
                }}
                value={get(v, valueField, "")}
                shouldDebounce
              />
              <Button
                intent={intentTypes.error}
                onClick={this.handleRemoveClick.bind(this, i)}
              >
                <X size={14} />
              </Button>
            </NestedKeyValueRow>
          )
        })}
        <KVFieldInput
          addValue={this.handleAddField}
          {...pick(this.props, [
            "field",
            "isKeyRequired",
            "isValueRequired",
            "keyField",
            "valueField",
          ])}
        />
      </Field>
    )
  }
}

UnwrappedQueryParamsKVField.propsTypes = {
  ...SHARED_KV_FIELD_PROPS,
  setQueryParams: PropTypes.func.isRequired,
  keyValueDelimiterChar: PropTypes.string.isRequired,
}

UnwrappedQueryParamsKVField.defaultProps = {
  ...SHARED_KV_FIELD_DEFAULT_PROPS,
  keyValueDelimiterChar: "|",
}

export default props => (
  <QueryParams>
    {({ queryParams, setQueryParams }) => {
      let values = get(queryParams, props.field, [])

      if (!isArray(values)) {
        values = [values]
      }

      return (
        <UnwrappedQueryParamsKVField
          {...props}
          setQueryParams={setQueryParams}
          values={values}
        />
      )
    }}
  </QueryParams>
)
