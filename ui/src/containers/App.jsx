import React, { Component } from 'react'
import { connect } from 'react-redux'
import { fetchDropdownOpts } from '../actions/'
import { ModalContainer } from '../components/'

export class App extends Component {
  componentDidMount() {
    const { dispatch } = this.props
    dispatch(fetchDropdownOpts())
  }
  componentDidUpdate(prevProps) {
    if (prevProps.location.pathname !== this.props.location.pathname) {
      window.scrollTo(0, 0)
    }
  }
  render() {
    const { uiGlobal: { modal } } = this.props
    return (
      <div className="app-root">
        {
          !!modal.visible && !!modal.modal ?
            <ModalContainer modal={modal.modal} /> : null
        }
        {this.props.children}
      </div>
    )
  }
}

function mapStateToProps(state) {
  return ({
    uiGlobal: state.uiGlobal
  })
}

export default connect(mapStateToProps)(App)
