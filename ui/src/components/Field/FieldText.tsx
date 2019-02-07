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
  validate?: (value: string) => boolean
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

export class FieldText extends React.PureComponent<IUnwrappedFieldTextProps> {
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

export class QueryParamsFieldText extends React.Component<
  IFieldTextProps,
  { error?: any }
> {
  state = {
    error: false,
  }
  render() {
    const { validate, name } = this.props
    return (
      <QueryParams>
        {({ queryParams, setQueryParams }) => (
          <FieldText
            {...this.props}
            error={this.state.error}
            value={get(queryParams, this.props.name, "")}
            onChange={value => {
              if (!validate || (validate && validate(value))) {
                setQueryParams(
                  {
                    [this.props.name]: value,
                  },
                  false
                )
              } else {
                this.setState({ error: "Invalid value." })
              }
            }}
          />
        )}
      </QueryParams>
    )
  }
}

export class FormikFieldText extends React.PureComponent<
  IFieldTextProps & FieldProps
> {
  render() {
    return (
      <FieldText
        {...omit(this.props, ["field", "form"])}
        onChange={value => {
          this.props.form.setFieldValue(this.props.field.name, value)
        }}
      />
    )
  }
}
