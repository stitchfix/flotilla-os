import * as React from "react"
import { X } from "react-feather"
import { get, pick, isArray } from "lodash"
import { FieldText } from "./FieldText"
import Button from "../styled/Button"
import NestedKeyValueRow from "../styled/NestedKeyValueRow"
import KVFieldInput from "./KVFieldInput"
import QueryParams from "../QueryParams/QueryParams"
import KVFieldContainer from "./KVFieldContainer"
import { intents } from "../../.."

interface IUnwrappedQueryParamsKVFieldProps {
  description?: string
  isKeyRequired: boolean
  isValueRequired: boolean
  keyField: string
  label: string
  name: string
  value: any[]
  valueField: string
  keyValueDelimiterChar: string
}

interface IQueryParamsKVFieldProps extends IUnwrappedQueryParamsKVFieldProps {
  queryParams: any
  setQueryParams: (query: object, shouldReplace: boolean) => void
}

type ParsedKVField = any

class UnwrappedQueryParamsKVField extends React.Component<
  IQueryParamsKVFieldProps
> {
  static defaultProps: Partial<IQueryParamsKVFieldProps> = {
    keyValueDelimiterChar: "|",
  }

  /** Handles input events for key fields. */
  handleKeyChange = (value: any, index: number): void => {
    const { keyField } = this.props

    this.handleChange({
      key: keyField,
      value,
      index,
    })
  }

  /** Handles input events for value fields. */
  handleValueChange = (value: any, index: number): void => {
    const { valueField } = this.props

    this.handleChange({
      key: valueField,
      value,
      index,
    })
  }

  /** Injects a new value into the values array prop then calls this.setValues. */
  handleChange = ({
    key,
    value,
    index,
  }: {
    key: string
    value: any
    index: number
  }): void => {
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
  handleAddField = (_: any, kv: ParsedKVField): void => {
    const values = this.getValues()
    this.setValues([...values, kv])
  }

  /** Removes a value specified by index. */
  handleRemoveClick = (index: number): void => {
    const values = this.getValues()
    this.setValues([...values.slice(0, index), ...values.slice(index + 1)])
  }

  /**
   * Stringifies a key value object (e.g. { name: "", value: ""}) into a string
   * delimited by the keyValueDelimiterChar prop (e.g. "key|value").
   */
  stringifyValue = (obj: ParsedKVField): string => {
    const { keyField, valueField, keyValueDelimiterChar } = this.props
    return `${obj[keyField]}${keyValueDelimiterChar}${obj[valueField]}`
  }

  /**
   * Parses a key value string object and transforms it into an object.
   */
  parseValue = (str: string): ParsedKVField => {
    const { keyField, valueField, keyValueDelimiterChar } = this.props
    const split = str.split(keyValueDelimiterChar)
    return {
      [keyField]: split[0],
      [valueField]: split[1],
    }
  }

  /** Calls props.setQueryParams to set new values. */
  setValues = (values: ParsedKVField[]): void => {
    const { setQueryParams, name } = this.props
    setQueryParams({ [name]: values.map(this.stringifyValue) }, false)
  }

  /** Transforms each value in the values prop to an object. */
  getValues = (): ParsedKVField[] => {
    const { value } = this.props

    return value.map(this.parseValue)
  }

  render() {
    const {
      name,
      label,
      keyField,
      isKeyRequired,
      isValueRequired,
      valueField,
      description,
    } = this.props

    return (
      <KVFieldContainer label={label} description={description}>
        {this.getValues().map((v, i) => {
          return (
            <NestedKeyValueRow key={i}>
              <FieldText
                name={keyField}
                isRequired={isKeyRequired}
                onChange={value => {
                  this.handleKeyChange(value, i)
                }}
                value={get(v, keyField, "")}
                shouldDebounce
              />
              <FieldText
                name={valueField}
                isRequired={isValueRequired}
                onChange={value => {
                  this.handleValueChange(value, i)
                }}
                value={get(v, valueField, "")}
                shouldDebounce
              />
              <Button
                intent={intents.ERROR}
                onClick={this.handleRemoveClick.bind(this, i)}
                type="button"
              >
                <X size={14} />
              </Button>
            </NestedKeyValueRow>
          )
        })}
        <KVFieldInput
          addValue={this.handleAddField}
          name={name}
          isKeyRequired={isKeyRequired}
          isValueRequired={isValueRequired}
          keyField={keyField}
          valueField={valueField}
        />
      </KVFieldContainer>
    )
  }
}

const WrappedQueryParamsKVField: React.SFC<
  IUnwrappedQueryParamsKVFieldProps
> = props => (
  <QueryParams>
    {({ queryParams, setQueryParams }) => {
      let value = get(queryParams, props.name, [])

      if (!isArray(value)) {
        value = [value]
      }

      return (
        <UnwrappedQueryParamsKVField
          {...props}
          setQueryParams={setQueryParams}
          value={value}
        />
      )
    }}
  </QueryParams>
)

export default WrappedQueryParamsKVField
