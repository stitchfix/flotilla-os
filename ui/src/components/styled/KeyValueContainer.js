import React, { Component } from "react"
import PropTypes from "prop-types"
import { ChevronUp, ChevronDown } from "react-feather"
import Button from "./Button"
import Card from "./Card"

export default class KeyValueContainer extends Component {
  static propTypes = {
    children: PropTypes.func,
    header: PropTypes.string,
  }
  constructor(props) {
    super(props)
    this.handleCollapseButtonClick = this.handleCollapseButtonClick.bind(this)
    this.handleJsonButtonClick = this.handleJsonButtonClick.bind(this)
    this.getState = this.getState.bind(this)
  }
  state = {
    collapsed: false,
    json: false,
  }
  handleJsonButtonClick() {
    this.setState(state => ({
      json: !state.json,
      collapsed: false,
    }))
  }
  handleCollapseButtonClick() {
    this.setState(state => ({ collapsed: !state.collapsed }))
  }
  getState() {
    return this.state
  }
  render() {
    const { json, collapsed } = this.state
    const { header, children } = this.props
    return (
      <Card
        title={header}
        actions={[
          <Button onClick={this.handleJsonButtonClick}>
            {!!json ? "Normal View" : "JSON View"}
          </Button>,
          <Button onClick={this.handleCollapseButtonClick}>
            {!!collapsed ? <ChevronDown size={14} /> : <ChevronUp size={14} />}
          </Button>,
        ]}
      >
        {children(this.getState())}
      </Card>
    )
  }
}
