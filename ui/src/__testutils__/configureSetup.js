import React from "react"
import { MemoryRouter } from "react-router-dom"
import { Provider } from "react-redux"
import { reduxForm } from "redux-form"
import enzyme from "enzyme"
import configureMockStore from "redux-mock-store"
import thunk from "redux-thunk"

//
// Set up an empty mock store.
//
const middlewares = [thunk]
const mockStore = configureMockStore(middlewares)

//
// configureSetup returns a `setup` function that can be invoked with a variety
// of options to simplify writing tests.
//
const configureSetup = (opts = {}) => {
  const {
    // A set of base props you want to pass your component.
    baseProps = {},

    // The component that is `connected` to Redux (via react-redux's `connect`
    // HOC) and usually the default export.
    connected,

    // The unconnected component. In addition to the component's default
    // export, you also have the option to just export the class itself. This
    // will make it easier to test component functionality not tied to Redux.
    unconnected,
  } = opts

  // Calling `configureSetup` will return this function, which you can call
  // again to return an Enzyme-mounted component.
  const setup = (setupOpts = {}) => {
    const {
      shallow = false,
      // Any specific props you want to pass to your component during testing.
      // This will be merged with the `baseProps` value provided to
      // configureSetup.
      props = {},

      // If your component is rendered by react-router's <Route> component, or
      // if you're wrapping your component with the `withRouter` HOC, the
      // component will be mounted inside a <MemoryRouter> component in Enzyme.
      // `routerProps` allows you to pass props (e.g. `initialEntries`) to the
      // MemoryRouter component.
      connectToRouter = false,
      routerProps = {},

      // Setting connectToRedux to true will wrap your component inside
      // react-redux's <Provider> component. By default, an empty mock store
      // will be provided but you can specify your own.
      connectToRedux = false,
      store = mockStore({}),

      // If your component is connected to redux-form, setting this to true
      // will wrap your component with the reduxForm HOC.
      connectToReduxForm = false,
      formName = "",
    } = setupOpts

    //
    // Merge opts.baseProps and setupOpts.props.
    //
    const mergedProps = {
      ...baseProps,
      ...props,
    }

    //
    // Determine what to mount.
    //
    let ToMount = unconnected

    if (connectToRedux) {
      ToMount = connected
    }

    if (connectToReduxForm) {
      ToMount = reduxForm({ formName })(unconnected)
    }

    //
    // Determine the mounting strategy for the component.
    //
    const mountMethod = !!shallow ? "shallow" : "mount"
    const shouldConnectToRedux = connectToRedux || connectToReduxForm
    if (shouldConnectToRedux && connectToRouter) {
      return enzyme[mountMethod](
        <Provider store={store}>
          <MemoryRouter {...routerProps}>
            <ToMount {...mergedProps} />
          </MemoryRouter>
        </Provider>
      )
    } else if (shouldConnectToRedux) {
      return enzyme[mountMethod](
        <Provider store={store}>
          <ToMount {...mergedProps} />
        </Provider>
      )
    } else if (connectToRouter) {
      return enzyme[mountMethod](
        <MemoryRouter>
          <ToMount {...mergedProps} />
        </MemoryRouter>
      )
    } else {
      return enzyme[mountMethod](<ToMount {...mergedProps} />)
    }
  }

  return setup
}

export default configureSetup
