import * as React from "react"
import { DebounceInput } from "react-debounce-input"
import { FieldProps } from "formik"
import { get, omit } from "lodash"
import StyledField from "../styled/Field"
import { Input, Textarea } from "../styled/Inputs"
import QueryParams from "../QueryParams/QueryParams"

interface IFieldTextProps {
  description?: string
  error?: any
  name: string
  inputRef?: () => void
  isNumber: boolean
  isRequired: boolean
  isTextArea: boolean
  label?: string
  shouldDebounce: boolean
  validate: any
}

interface IUnwrappedFieldTextProps extends IFieldTextProps {
  onChange: (value: any) => void
  value: any
}

/** Props to be passed to various input components rendered by FieldText */
interface ISharedInputProps {
  value: any
  onChange: (evt: React.SyntheticEvent) => void
  ref?: () => void
}

class FieldText extends React.PureComponent<IUnwrappedFieldTextProps> {
  static defaultProps: Partial<IUnwrappedFieldTextProps> = {
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
      name,
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
      <StyledField
        label={label}
        isRequired={isRequired}
        description={description}
        error={error}
      >
        {input}
      </StyledField>
    )
  }
}

export const QueryParamsFieldText: React.SFC<IFieldTextProps> = props => {
  return (
    <QueryParams>
      {({ queryParams, setQueryParams }) => (
        <FieldText
          {...props}
          value={get(queryParams, props.name, "")}
          onChange={value => {
            setQueryParams(
              {
                [props.name]: value,
              },
              false
            )
          }}
        />
      )}
    </QueryParams>
  )
}

export const FormikFieldText: React.SFC<
  IFieldTextProps & FieldProps
> = props => {
  return (
    <FieldText
      {...omit(props, ["field", "form"])}
      onChange={value => {
        props.form.setFieldValue(props.field.name, value)
      }}
    />
  )
}
