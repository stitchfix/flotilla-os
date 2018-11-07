import React, { Component } from "react"
import PropTypes from "prop-types"
import { FieldArray } from "redux-form"
import { isFunction } from "lodash"
import { X } from "react-feather"

const ReduxFormGroupArrayPropTypes = PropTypes.shape({
  name: PropTypes.string.isRequired,
  label: PropTypes.string.isRequired,
  description: PropTypes.string,
  fieldDisplayHints: PropTypes.arrayOf(
    PropTypes.shape({
      label: PropTypes.string.isRequired,
      description: PropTypes.string,
    })
  ),
})

class ReduxFormGroupArray extends Component {
  static propTypes = {
    fields: PropTypes.any,
    custom: ReduxFormGroupArrayPropTypes,
  }
  constructor(props) {
    super(props)
    this.getRowProps = this.getRowProps.bind(this)
    this.getFieldProps = this.getFieldProps.bind(this)
  }
  getFieldProps(props) {
    const { field, name, flexWidth } = props
    let derivedName = field

    if (!!name) derivedName = `${field}.${name}`

    return {
      ...props,
      name: derivedName,
      style: { flex: flexWidth },
    }
  }
  getRowProps(props) {
    return {
      ...props,
      onRemoveClick: () => {
        this.props.fields.remove(props.index)
        this.props.custom.onRemoveField(props.index)
      },
    }
  }
  getStateAndHelpers() {
    return {
      getFieldProps: this.getFieldProps,
      getRowProps: this.getRowProps,
      fields: this.props.fields,
    }
  }
  render() {
    const { fields, custom } = this.props

    return (
      <div className="redux-form-group-array-container">
        <div className="redux-form-group-array-header">
          <div>
            <div>{custom.label}</div>
            <div className="form-group-helper-text">{custom.description}</div>
          </div>
          <button
            className="pl-button pl-intent-primary"
            type="button"
            onClick={() => {
              fields.push()

              if (isFunction(custom.onAddField)) {
                custom.onAddField()
              }
            }}
          >
            Add
          </button>
        </div>
        <div className="full-width">
          <div className="redux-form-group-array-row">
            {custom.fieldDisplayHints &&
              custom.fieldDisplayHints.map((field, i) => (
                <div key={i} style={{ flex: field.flexWidth }}>
                  <div>{field.label}</div>
                  <div className="pl-form-group-helper-text">
                    {field.description}
                  </div>
                </div>
              ))}
            {/* An invisible button to help alignment. */}
            <button
              className="pl-button"
              type="button"
              disabled
              style={{ opacity: 0 }}
            >
              <X size={14} />
            </button>
          </div>
          {custom.render(this.getStateAndHelpers())}
        </div>
      </div>
    )
  }
}

const asFieldArray = UnwrappedFieldArray => {
  const WrappedFieldArray = props => (
    <FieldArray
      name={props.name}
      component={UnwrappedFieldArray}
      props={{
        custom: {
          name: props.name,
          label: props.label,
          description: props.description,
          render: props.render,
          fieldDisplayHints: props.fieldDisplayHints,
          onAddField: props.onAddField,
          onRemoveField: props.onRemoveField,
        },
      }}
    />
  )
  WrappedFieldArray.propTypes = ReduxFormGroupArrayPropTypes
  return WrappedFieldArray
}

export default asFieldArray(ReduxFormGroupArray)
