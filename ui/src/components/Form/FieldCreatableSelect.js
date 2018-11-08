import React from "react"
import PropTypes from "prop-types"
import CreatableSelect from "react-select/lib/Creatable"
import { asField } from "informed"
import { get } from "lodash"
import Field from "./Field"

const stringToSelectOpt = (str = "") => ({ label: str, value: str })
const selectOptToString = opt => get(opt, "value", "")

const FieldCreatableSelect = asField(({ fieldState, fieldApi, ...props }) => (
  <Field label={props.label}>
    <CreatableSelect
      isClearable
      options={props.options}
      value={stringToSelectOpt(fieldState.value)}
      // Note: an empty `onInputChange` prop will prevent CreatableSelect from
      // crashing.
      onInputChange={() => {}}
      onChange={option => {
        fieldApi.setValue(selectOptToString(option))
      }}
      isMulti={props.isMulti}
    />
  </Field>
))

FieldCreatableSelect.propTypes = {
  isMulti: PropTypes.bool.isRequired,
  label: PropTypes.string,
  options: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string,
      value: PropTypes.string,
    })
  ),
}

FieldCreatableSelect.defaultProps = {
  isMulti: false,
  options: [],
}

export default FieldCreatableSelect
