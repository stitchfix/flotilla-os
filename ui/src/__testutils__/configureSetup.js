import React from "react"
import { MemoryRouter } from "react-router-dom"
import { Provider } from "react-redux"
import { reduxForm } from "redux-form"
import { mount } from "enzyme"
import configureMockStore from "redux-mock-store"
import thunk from "redux-thunk"

// Set up an empty mock store.
const middlewares = [thunk]
const mockStore = configureMockStore(middlewares)

export default function configureSetup(opts = {}) {
  const { baseProps = {}, connected, unconnected } = opts

  return ({
    props = {},
    routerProps = {},
    connectToRedux = false,
    connectToRouter = false,
    connectToReduxForm = false,
    formName = "",
    store = mockStore({}),
  } = {}) => {
    // Merge base props with props.
    const mergedProps = {
      ...baseProps,
      ...props,
    }

    let ToMount = unconnected

    if (connectToRedux) {
      ToMount = connected
    }

    if (connectToReduxForm) {
      ToMount = reduxForm({ formName })(ToMount)
    }

    if (connectToRedux && connectToRouter) {
      return mount(
        <Provider store={store}>
          <MemoryRouter {...routerProps}>
            <ToMount {...mergedProps} />
          </MemoryRouter>
        </Provider>
      )
    } else if (connectToRedux || connectToReduxForm) {
      return mount(
        <Provider store={store}>
          <ToMount {...mergedProps} />
        </Provider>
      )
    } else if (connectToRouter) {
      return mount(
        <MemoryRouter>
          <ToMount {...mergedProps} />
        </MemoryRouter>
      )
    } else {
      return mount(<ToMount {...mergedProps} />)
    }
  }
}
