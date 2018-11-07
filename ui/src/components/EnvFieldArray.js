import React from "react"
import PropTypes from "prop-types"
import ReduxFormGroupInput from "./ReduxFormGroupInput"
import ReduxFormGroupArray from "./ReduxFormGroupArray"
import ReduxFormGroupArrayRow from "./ReduxFormGroupArrayRow"

const EnvFieldArray = props => {
  const { handleEnvCreate, handleEnvUpdate, handleEnvRemove } = props

  return (
    <ReduxFormGroupArray
      {...props}
      name="env"
      label="Environment Variables"
      fieldDisplayHints={[
        { label: "Name", flexWidth: 1 },
        { label: "Value", flexWidth: 1 },
      ]}
      onAddField={handleEnvCreate}
      onRemoveField={handleEnvRemove}
      render={({ fields, getRowProps, getFieldProps }) => (
        <div className="flex ff-cn">
          {fields.map((field, i) => (
            <ReduxFormGroupArrayRow {...getRowProps({ index: i })} key={i}>
              <ReduxFormGroupInput
                isRequired
                {...getFieldProps({
                  index: i,
                  name: "name",
                  field,
                  flexWidth: 1,
                  onChange: (evt, newVal) => {
                    handleEnvUpdate({
                      nameOrValue: "name",
                      index: i,
                      value: newVal,
                    })
                  },
                })}
              />
              <ReduxFormGroupInput
                isRequired
                {...getFieldProps({
                  index: i,
                  name: "value",
                  field,
                  flexWidth: 1,
                  onChange: (evt, newVal) => {
                    handleEnvUpdate({
                      nameOrValue: "value",
                      index: i,
                      value: newVal,
                    })
                  },
                })}
              />
            </ReduxFormGroupArrayRow>
          ))}
        </div>
      )}
    />
  )
}

EnvFieldArray.propTypes = {
  handleEnvCreate: PropTypes.func,
  handleEnvUpdate: PropTypes.func,
  handleEnvRemove: PropTypes.func,
}

EnvFieldArray.defaultProps = {
  handleEnvCreate: () => {},
  handleEnvUpdate: () => {},
  handleEnvRemove: () => {},
}

export default EnvFieldArray
