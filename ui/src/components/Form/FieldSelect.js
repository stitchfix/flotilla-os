import React from "react"
import PropTypes from "prop-types"
import Select from "react-select"
import { asField } from "informed"
import Field from "./Field"

const stringToSelectOpt = str => ({ label: str, value: str })
const selectOptToString = opt => opt.value

const FieldSelect = asField(({ fieldState, fieldApi, ...props }) => (
  <Field label={props.label}>
    <Select
      options={props.options}
      value={stringToSelectOpt(fieldState.value)}
      onChange={option => {
        fieldApi.setValue(selectOptToString(option))
      }}
      isMulti={props.isMulti}
    />
  </Field>
))

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
