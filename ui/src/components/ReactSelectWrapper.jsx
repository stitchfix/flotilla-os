import React, { Component } from 'react'
import PropTypes from 'prop-types'
import Select from 'react-select'
import { isEqual } from 'lodash'

export default class ReactSelectWrapper extends Component {
  static propTypes = {
    allowCreate: PropTypes.bool,
    onChange: PropTypes.func,
    multi: PropTypes.bool,
    value: PropTypes.any,
    options: PropTypes.arrayOf(PropTypes.shape({
      label: PropTypes.string,
      value: PropTypes.string,
    }))
  }
  static defaultProps = {
    allowCreate: false,
    multi: false,
  }
  constructor(props) {
    super(props)
    this.handleChange = this.handleChange.bind(this)
  }
  state = {
    options: []
  }
  componentDidMount() {
    // [IMPORTANT] Due to issues with React Select, this step is
    // required in order for `Select.Creatable` to work.
    this.setState({ options: this.props.options })
  }
  componentDidUpdate(prevProps) {
    // This is to handle a Select component receiving new options.
    // E.g. image -> image tag select.
    if (!isEqual(prevProps.options, this.props.options)) {
      this.setState({ options: this.props.options })
    }
  }
  handleChange(option) {
    // Clearing the select component will call `handleChange` with `null`
    if (!!option) {
      // Differentiate between single/multi values
      if (Array.isArray(option)) {
        this.props.onChange(option.map(o => o.value))
      } else {
        this.props.onChange(option.value)
      }
    } else {
      this.props.onChange(null)
    }
  }
  render() {
    const { allowCreate, multi, value, onFocus, onBlur } = this.props
    const { options } = this.state

    let _value = value

    if (multi) {
      if (Array.isArray(value)) {
        _value = value.filter(val => !!val)
      } else {
        _value = value === '' ? null : value
      }
    }

    if (allowCreate) {
      return (
        <Select.Creatable
          clearable
          multi={multi}
          onChange={this.handleChange}
          options={options}
          value={_value}
          onBlur={onBlur}
          onFocus={onFocus}
        />
      )
    } else {
      return (
        <Select
          clearable
          multi={multi}
          onChange={this.handleChange}
          options={options}
          value={_value}
          onBlur={onBlur}
          onFocus={onFocus}
        />
      )
    }
  }
}
