import React, { PureComponent } from "react"
import PropTypes from "prop-types"
import Select from "react-select"
import CreatableSelect from "react-select/lib/Creatable"
import { Field as RFField } from "react-form"
import { get, isArray, isString, isEmpty, isFunction } from "lodash"
import Field from "../styled/Field"
import {
  stringToSelectOpt,
  selectOptToString,
  selectTheme,
  selectStyles,
} from "../../utils/reactSelectHelpers"
import * as requestStateTypes from "../../constants/requestStateTypes"
import PopupContext from "../Popup/PopupContext"
import QueryParams from "../QueryParams/QueryParams"

class FieldSelect extends PureComponent {
  state = {
    requestState: requestStateTypes.NOT_READY,
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

  requestOptions = () => {
    const { requestOptionsFn, getOptions, renderPopup } = this.props

    this.setState({ inFlight: true, error: false })

    requestOptionsFn()
      .then(res => {
        this.setState({
          options: getOptions(res),
          inFlight: false,
          requestState: requestStateTypes.READY,
          error: false,
        })
      })
      .catch(error => {
        this.setState({
          inFlight: false,
          requestState: requestStateTypes.ERROR,
          error,
        })
        renderPopup({
          body: error.toString(),
          title: "An error occurred.",
        })
      })
  }

  getSharedProps = () => {
    const { isMulti, options, shouldRequestOptions } = this.props

    return {
      closeMenuOnSelect: !isMulti,
      isClearable: true,
      isMulti: isMulti,
      onChange: selected => {
        this.handleSelectChange(selected)
      },
      options: shouldRequestOptions ? this.state.options : options,
      styles: selectStyles,
      theme: selectTheme,
      value: this.getValue(),
    }
  }

  getValue = () => {
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

  handleSelectChange = selected => {
    const { isMulti, onChange } = this.props

    if (selected === null) {
      if (isMulti) {
        onChange([])
      } else {
        onChange("")
      }

      return
    }

    if (isMulti) {
      onChange(selected.map(selectOptToString))
      return
    }

    onChange(selected.value)
  }

  isReady = () => {
    const { shouldRequestOptions } = this.props
    const { requestState } = this.state

    if (shouldRequestOptions && requestState !== requestStateTypes.READY) {
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

FieldSelect.propTypes = {
  field: PropTypes.string.isRequired,
  getOptions: PropTypes.func,
  isCreatable: PropTypes.bool.isRequired,
  isMulti: PropTypes.bool.isRequired,
  label: PropTypes.string,
  options: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string,
      value: PropTypes.string,
    })
  ),
  renderPopup: PropTypes.func,
  requestOptionsFn: PropTypes.func,
  shouldRequestOptions: PropTypes.bool.isRequired,
}

FieldSelect.defaultProps = {
  getOptions: res => res,
  isCreatable: false,
  isMulti: false,
  options: [],
  shouldRequestOptions: false,
}

const FieldSelectWithPopupContext = props => (
  <PopupContext.Consumer>
    {ctx => <FieldSelect {...props} renderPopup={ctx.renderPopup} />}
  </PopupContext.Consumer>
)

export const QueryParamsFieldSelect = props => (
  <QueryParams>
    {({ queryParams, setQueryParams }) => (
      <FieldSelectWithPopupContext
        {...props}
        value={get(queryParams, props.field, "")}
        onChange={value => {
          setQueryParams({
            [props.field]: value,
          })
        }}
      />
    )}
  </QueryParams>
)

export const ReactFormFieldSelect = props => (
  <RFField field={props.field} validate={props.validate}>
    {fieldAPI => (
      <FieldSelectWithPopupContext
        {...props}
        value={get(fieldAPI, "value", "")}
        onChange={value => {
          fieldAPI.setValue(value)
        }}
        error={get(fieldAPI, "error", null)}
      />
    )}
  </RFField>
)
