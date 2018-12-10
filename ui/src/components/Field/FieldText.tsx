import * as React from "react"
import DebounceInput from "react-debounce-input"
import { Field as RFField } from "react-form"
import { get, has } from "lodash"
import Field from "../styled/Field"
import { Input, Textarea } from "../styled/Inputs"
import QueryParams from "../QueryParams/QueryParams"

interface IFieldTextProps {
  description?: string
  error?: any
  field: string
  inputRef?: () => void
  isNumber: boolean
  isRequired: boolean
  isTextArea: boolean
  label?: string
  onChange: (value: any) => void
  shouldDebounce: boolean
  value: any
}

/** Props to be passed to various input components rendered by FieldText */
interface ISharedInputProps {
  value: any
  onChange: (evt: React.SyntheticEvent) => void
  ref?: () => void
}

class FieldText extends React.PureComponent<IFieldTextProps> {
  static defaultProps: Partial<IFieldTextProps> = {
    isNumber: false,
    isRequired: false,
    isTextArea: false,
    shouldDebounce: false,
  }

  render() {
    const {
      description,
      error,
      inputRef,
      isNumber,
      isRequired,
      isTextArea,
      label,
      onChange,
      shouldDebounce,
      value,
      field,
    } = this.props

    // Common props for all input components
    let sharedProps: ISharedInputProps = {
      value,
      onChange: (evt: React.SyntheticEvent) => {
        // Per https://stackoverflow.com/a/42084103
        const target = evt.target as HTMLInputElement
        onChange(target.value)
      },
    }

    if (!!inputRef) {
      sharedProps.ref = inputRef
    }

    // Assign input element based on various props.
    let input

    if (isTextArea) {
      input = <Textarea {...sharedProps} />
    } else if (shouldDebounce) {
      input = (
        <DebounceInput
          {...sharedProps}
          element={Input}
          debounceTimeout={250}
          minLength={1}
          type={isNumber ? "number" : "text"}
        />
      )
    } else {
      input = <Input type={isNumber ? "number" : "text"} {...sharedProps} />
    }

    return (
      <Field
        label={label}
        isRequired={isRequired}
        description={description}
        error={error}
      >
        {input}
      </Field>
    )
  }
}

export const QueryParamsFieldText = props => {
  return (
    <QueryParams>
      {({ queryParams, setQueryParams }) => (
        <FieldText
          {...props}
          value={get(queryParams, props.field, "")}
          onChange={value => {
            setQueryParams(
              {
                [props.field]: value,
              },
              false
            )
          }}
        />
      )}
    </QueryParams>
  )
}

export const ReactFormFieldText = props => {
  return (
    <RFField field={props.field} validate={props.validate}>
      {fieldAPI => {
        return (
          <FieldText
            {...props}
            value={get(fieldAPI, "value", "")}
            onChange={value => fieldAPI.setValue(value)}
            error={get(fieldAPI, "error", null)}
          />
        )
      }}
    </RFField>
  )
}
