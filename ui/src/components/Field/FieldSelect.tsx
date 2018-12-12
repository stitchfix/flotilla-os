import * as React from "react"
import Select from "react-select"
import CreatableSelect from "react-select/lib/Creatable"
import { get, isArray, isString, isEmpty, isFunction, omit } from "lodash"
import Field from "../styled/Field"
import {
  stringToSelectOpt,
  selectOptToString,
  selectTheme,
  selectStyles,
} from "../../helpers/reactSelectHelpers"
import PopupContext from "../Popup/PopupContext"
import QueryParams from "../QueryParams/QueryParams"
import { requestStates, IReactSelectOption, IPopupProps } from "../../.."
import { ValueType, ActionMeta } from "react-select/lib/types"
import { FieldProps } from "formik"

interface IFieldSelectProps {
  description?: string
  error: any
  name: string
  getOptions?: (res: any) => IReactSelectOption[]
  isCreatable: boolean
  isMulti: boolean
  isRequired: boolean
  label?: string
  onChange: (value: string | string[]) => void
  options?: IReactSelectOption[]
  requestOptionsFn?: () => Promise<any>
  shouldRequestOptions: boolean
  value: any
}

interface IUnwrappedFieldSelectFieldSelectProps extends IFieldSelectProps {
  renderPopup: (popupProps: IPopupProps) => void
}

interface IFieldSelectState {
  requestState: requestStates
  inFlight: boolean
  options: IReactSelectOption[]
  error: any
}

interface ISharedSelectProps {
  name: string
  closeMenuOnSelect: boolean
  isClearable: boolean
  isMulti: boolean
  onChange: (
    selected: ValueType<IReactSelectOption | IReactSelectOption[]>,
    action: ActionMeta
  ) => void
  options: IReactSelectOption[] | undefined
  styles: any
  theme: any
  value: any
}

class UnwrappedFieldSelect extends React.PureComponent<
  IUnwrappedFieldSelectFieldSelectProps,
  IFieldSelectState
> {
  static defaultProps: Partial<IUnwrappedFieldSelectFieldSelectProps> = {
    getOptions: res => res,
    isCreatable: false,
    isMulti: false,
    options: [],
    shouldRequestOptions: false,
  }

  state = {
    requestState: requestStates.NOT_READY,
    inFlight: false,
    options: [],
    error: false,
  }

  componentDidMount() {
    const { shouldRequestOptions, requestOptionsFn } = this.props

    if (shouldRequestOptions && isFunction(requestOptionsFn)) {
      this.requestOptions()
    }
  }

  requestOptions = (): void => {
    const { requestOptionsFn, getOptions, renderPopup } = this.props

    this.setState({ inFlight: true, error: false })

    if (!!requestOptionsFn) {
      requestOptionsFn()
        .then(res => {
          let options = res

          if (!!getOptions) {
            options = getOptions(res)
          }
          this.setState({
            options,
            inFlight: false,
            requestState: requestStates.READY,
            error: false,
          })
        })
        .catch(error => {
          this.setState({
            inFlight: false,
            requestState: requestStates.ERROR,
            error,
          })
          renderPopup({
            body: error.toString(),
            title: "An error occurred while fetching select options",
          })
        })
    }
  }

  getSharedProps = (): ISharedSelectProps => {
    const { isMulti, name, options, shouldRequestOptions } = this.props

    return {
      closeMenuOnSelect: !isMulti,
      isClearable: true,
      isMulti: isMulti,
      name,
      onChange: (
        selected: ValueType<IReactSelectOption | IReactSelectOption[]>
      ) => {
        this.handleSelectChange(selected)
      },
      options: !!shouldRequestOptions ? this.state.options : options,
      styles: selectStyles,
      theme: selectTheme,
      value: this.getValue(),
    }
  }

  getValue = (): IReactSelectOption | IReactSelectOption[] => {
    const { isMulti, value } = this.props

    if (isMulti) {
      if (isArray(value)) {
        return value.map(stringToSelectOpt)
      } else if (isString(value) && !isEmpty(value)) {
        return [stringToSelectOpt(value)]
      } else {
        return []
      }
    }

    return stringToSelectOpt(value)
  }

  handleSelectChange = (
    selected: ValueType<IReactSelectOption | IReactSelectOption[]>
  ) => {
    const { isMulti, onChange } = this.props

    if (selected === null || selected === undefined) {
      if (isMulti) {
        onChange([])
      } else {
        onChange("")
      }

      return
    }

    if (isMulti) {
      onChange((selected as IReactSelectOption[]).map(selectOptToString))
      return
    }

    onChange((selected as IReactSelectOption).value)
  }

  isReady = (): boolean => {
    const { shouldRequestOptions } = this.props
    const { requestState } = this.state

    if (shouldRequestOptions && requestState !== requestStates.READY) {
      return false
    }

    return true
  }

  render() {
    const { error, isCreatable, label, isRequired, description } = this.props

    if (!this.isReady()) {
      return <span />
    }

    const sharedProps = this.getSharedProps()
    let select

    if (isCreatable) {
      select = <CreatableSelect {...sharedProps} onInputChange={() => {}} />
    } else {
      select = <Select {...sharedProps} />
    }

    return (
      <Field
        label={label}
        isRequired={isRequired}
        description={description}
        error={error}
      >
        {select}
      </Field>
    )
  }
}

const FieldSelectWithPopupContext: React.SFC<IFieldSelectProps> = props => (
  <PopupContext.Consumer>
    {ctx => <UnwrappedFieldSelect {...props} renderPopup={ctx.renderPopup} />}
  </PopupContext.Consumer>
)

export const QueryParamsFieldSelect: React.SFC<
  IUnwrappedFieldSelectFieldSelectProps
> = props => (
  <QueryParams>
    {({ queryParams, setQueryParams }) => (
      <FieldSelectWithPopupContext
        {...props}
        value={get(queryParams, props.name, "")}
        onChange={(value: string | string[]) => {
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

/**
 * This component is designed to be passed to a Formik.Field component's
 * `component` prop. E.g.
 * <Formik.Field component={FormikFieldSelect} options={[]} />
 */
export const FormikFieldSelect: React.SFC<
  IUnwrappedFieldSelectFieldSelectProps & FieldProps
> = props => {
  return (
    <UnwrappedFieldSelect
      {...omit(props, ["field", "form"])}
      onChange={value => {
        props.form.setFieldValue(props.field.name, value)
      }}
    />
  )
}
