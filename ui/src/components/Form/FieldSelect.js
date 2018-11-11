import React, { Component } from "react"
import PropTypes from "prop-types"
import Select from "react-select"
import CreatableSelect from "react-select/lib/Creatable"
import { get, isArray, isString, isEmpty } from "lodash"
import { Field as RFField } from "react-form"
import Field from "../styled/Field"
import {
  stringToSelectOpt,
  selectOptToString,
  selectTheme,
  selectStyles,
} from "../../utils/reactSelectHelpers"

class FieldSelect extends Component {
  getSharedProps = fieldAPI => {
    const { isMulti, options } = this.props

    return {
      closeMenuOnSelect: !isMulti,
      isClearable: true,
      isMulti: isMulti,
      onChange: selected => {
        this.handleSelectChange(selected, fieldAPI)
      },
      options: options,
      styles: selectStyles,
      theme: selectTheme,
      value: this.getValue(fieldAPI),
    }
  }

  getValue = fieldAPI => {
    const { isMulti } = this.props
    const value = get(fieldAPI, "value")

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

  handleSelectChange = (selected, fieldAPI) => {
    const { isMulti } = this.props

    if (isMulti) {
      fieldAPI.setValue(selected.map(selectOptToString))
      return
    }

    fieldAPI.setValue(selected.value)
  }

  render() {
    const { field, isCreatable, label, isRequired, description } = this.props
    return (
      <RFField field={field}>
        {fieldAPI => {
          const sharedProps = this.getSharedProps(fieldAPI)
          let select

          if (isCreatable) {
            select = (
              <CreatableSelect {...sharedProps} onInputChange={() => {}} />
            )
          } else {
            select = <Select {...sharedProps} />
          }

          return (
            <Field
              label={label}
              isRequired={isRequired}
              description={description}
              error={fieldAPI.error}
            >
              {select}
            </Field>
          )
        }}
      </RFField>
    )
  }
}

FieldSelect.propTypes = {
  isMulti: PropTypes.bool.isRequired,
  label: PropTypes.string,
  options: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string,
      value: PropTypes.string,
    })
  ),
}

FieldSelect.defaultProps = {
  isMulti: false,
  options: [],
}

export default FieldSelect
