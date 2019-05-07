import * as React from "react"
import Select from "react-select"
import CreatableSelect from "react-select/lib/Creatable"
import { isArray, isString, isEmpty, omit, Omit } from "lodash"
import Field from "../styled/Field"
import {
  stringToSelectOpt,
  selectOptToString,
  selectTheme,
  selectStyles,
} from "../../helpers/reactSelectHelpers"
import { flotillaUIRequestStates, IReactSelectOption } from "../../types"
import { ValueType, ActionMeta } from "react-select/lib/types"
import Request from "../Request/Request"

export interface IProps
  extends Omit<IConnectedProps, "requestOptionsFn" | "getOptionsFromResponse"> {
  requestState: flotillaUIRequestStates
  inFlight: boolean
  error: any
  request: (args?: any) => void
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

/**
 * Provides several convenient defaults around react-select.
 */
export class ReactSelectWrapper extends React.PureComponent<IProps> {
  static defaultProps = {
    isCreatable: false,
    isMulti: false,
    options: [],
    shouldRequestOptions: false,
  }

  componentDidMount() {
    const { shouldRequestOptions, request } = this.props

    // Make an API call to fetch the options if necessary.
    if (shouldRequestOptions) {
      request()
    }
  }

  getSharedProps = (): ISharedSelectProps => {
    const { isMulti, name, options } = this.props

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
      options,
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
    const { shouldRequestOptions, requestState } = this.props

    if (
      shouldRequestOptions &&
      requestState !== flotillaUIRequestStates.READY
    ) {
      return false
    }

    return true
  }

  render() {
    if (!this.isReady()) {
      return <span />
    }

    const sharedProps = this.getSharedProps()

    if (this.props.isCreatable) {
      return <CreatableSelect {...sharedProps} onInputChange={() => {}} />
    } else {
      return <Select {...sharedProps} />
    }
  }
}

interface IConnectedProps {
  name: string
  isCreatable: boolean
  isMulti: boolean
  onChange: (value: string | string[]) => void
  options?: IReactSelectOption[]
  requestOptionsFn?: () => Promise<any>
  shouldRequestOptions: boolean
  value: any
  onRequestError?: (e: any) => void
  getOptionsFromResponse: (data: any) => IReactSelectOption[]
}

class ConnectedReactSelectWrapper extends React.Component<IConnectedProps> {
  static defaultProps = {
    shouldRequestOptions: false,
    getOptionsFromResponse: (res: any): any => res,
    isCreatable: false,
    isMulti: false,
  }

  render() {
    const {
      requestOptionsFn,
      shouldRequestOptions,
      getOptionsFromResponse,
      options,
    } = this.props
    return (
      <Request requestFn={requestOptionsFn} shouldRequestOnMount={false}>
        {requestChildProps => (
          <ReactSelectWrapper
            {...this.props}
            {...omit(requestChildProps, "data")}
            options={
              shouldRequestOptions
                ? getOptionsFromResponse(requestChildProps.data)
                : options
            }
          />
        )}
      </Request>
    )
  }
}

export default ConnectedReactSelectWrapper
