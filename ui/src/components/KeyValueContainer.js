import React, { Component } from "react"
import PropTypes from "prop-types"
import { ChevronUp, ChevronDown } from "react-feather"
import { Button, Card, FormGroup } from "aa-ui-components"

export default class KeyValueContainer extends Component {
  static propTypes = {
    header: PropTypes.string,
    children: PropTypes.func,
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
    const { header, children, cardProps } = this.props
    return (
      <Card
        header={
          <div className="flex ff-rn j-sb a-c full-width">
            <div>{header}</div>
            <div className="flex ff-rn j-fs a-c with-horizontal-child-margin">
              <Button onClick={this.handleJsonButtonClick}>
                {!!json ? "Normal View" : "JSON View"}
              </Button>
              <Button onClick={this.handleCollapseButtonClick}>
                {!!collapsed ? (
                  <ChevronDown size={14} />
                ) : (
                  <ChevronUp size={14} />
                )}
              </Button>
            </div>
          </div>
        }
        collapsed={collapsed}
        contentStyle={{ padding: 0 }}
        {...cardProps}
      >
        {children(this.getState())}
      </Card>
    )
  }
}
