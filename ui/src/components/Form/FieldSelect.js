import React from "react"
import PropTypes from "prop-types"
import Select from "react-select"
import CreatableSelect from "react-select/lib/Creatable"
import { get, isString } from "lodash"
import { Field as RFField } from "react-form"
import Field from "./Field"

const stringToSelectOpt = (str = "") => {
  let ret = isString(str) ? str : ""
  return { label: ret, value: ret }
}
const selectOptToString = opt => get(opt, "value", "")

const FieldSelect = props => {
  return (
    <RFField field={props.field}>
      {fieldAPI => {
        const sharedProps = {
          closeMenuOnSelect: !props.isMulti,
          isClearable: true,
          isMulti: props.isMulti,
          options: props.options,
          value: props.isMulti
            ? get(fieldAPI, "value", []).map(stringToSelectOpt)
            : stringToSelectOpt(fieldAPI.value),
          onChange: selected => {
            if (props.isMulti) {
              fieldAPI.setValue(selected.map(selectOptToString))
              return
            }

            fieldAPI.setValue(selected.value)
          },
          styles: {
            container: provided => ({
              ...provided,
              width: "100%",
            }),
          },
        }
        let select

        if (props.isCreatable) {
          select = <CreatableSelect {...sharedProps} onInputChange={() => {}} />
        } else {
          select = <Select {...sharedProps} />
        }

        return (
          <Field
            label={props.label}
            isRequired={props.isRequired}
            description={props.description}
            error={fieldAPI.error}
          >
            {select}
          </Field>
        )
      }}
    </RFField>
  )
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
