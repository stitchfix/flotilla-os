import * as React from "react"
import { mount } from "enzyme"
import Request, { IProps, IChildProps } from "../Request"
import { flotillaUIRequestStates } from "../../../types"

const DEFAULT_PROPS: Partial<IProps> = {
  shouldRequestOnMount: true,
  requestFn: () => new Promise(resolve => {}),
  children: (props: IChildProps) => <div />,
}

describe("Request", () => {
  it("calls its request method on componentDidMount if the shouldRequestOnMount prop is true", () => {
    let realRequest = Request.prototype.request
    Request.prototype.request = jest.fn()
    const initialRequestArgs = { foo: "bar" }
    const wrapper = mount<Request>(
      <Request {...DEFAULT_PROPS} initialRequestArgs={initialRequestArgs} />
    )
    expect(Request.prototype.request).toHaveBeenCalledTimes(1)
    expect(Request.prototype.request).toHaveBeenCalledWith(initialRequestArgs)
    Request.prototype.request = realRequest
  })

  it("calls its children prop with the return value from its getChildProps method", () => {
    const children = jest.fn((props: IChildProps) => <div />)
    const wrapper = mount<Request>(
      <Request {...DEFAULT_PROPS} children={children} />
    )
    const childProps: IChildProps = {
      ...wrapper.state(),
      request: wrapper.instance().request,
    }
    expect(children).toHaveBeenCalledWith(childProps)
  })

  it("can request data", async () => {
    const data = { foo: "bar" }
    const requestFn = () =>
      new Promise(resolve => {
        resolve(data)
      })
    const wrapper = mount<Request>(
      <Request {...DEFAULT_PROPS} requestFn={requestFn} />
    )

    expect(wrapper.state()).toEqual({
      inFlight: true,
      data: null,
      requestState: flotillaUIRequestStates.NOT_READY,
      error: false,
    })

    setImmediate(() => {
      expect(wrapper.state()).toEqual({
        inFlight: false,
        data,
        requestState: flotillaUIRequestStates.READY,
        error: false,
      })
    })
  })

  it("can handle errors", async () => {
    const error = { foo: "bar" }
    const errorRequestFn = () =>
      new Promise((resolve, reject) => {
        reject(error)
      })
    const wrapper = mount<Request>(
      <Request {...DEFAULT_PROPS} requestFn={errorRequestFn} />
    )

    expect(wrapper.state()).toEqual({
      inFlight: true,
      data: null,
      requestState: flotillaUIRequestStates.NOT_READY,
      error: false,
    })

    setImmediate(() => {
      expect(wrapper.state()).toEqual({
        inFlight: false,
        data: null,
        requestState: flotillaUIRequestStates.ERROR,
        error,
      })
    })
  })

  it("can handle multiple requestFns", async () => {
    const one = { one: "one" }
    const two = { two: "two" }
    const three = { three: "three" }
    const requestFns = [
      () =>
        new Promise(r => {
          r(one)
        }),
      () =>
        new Promise(r => {
          r(two)
        }),
      () =>
        new Promise(r => {
          r(three)
        }),
    ]
    const wrapper = mount<Request>(
      <Request
        {...DEFAULT_PROPS}
        requestFn={requestFns}
        initialRequestArgs={[]}
      />
    )

    expect(wrapper.state()).toEqual({
      inFlight: true,
      data: null,
      requestState: flotillaUIRequestStates.NOT_READY,
      error: false,
    })

    setImmediate(() => {
      expect(wrapper.state()).toEqual({
        inFlight: false,
        data: [one, two, three],
        requestState: flotillaUIRequestStates.READY,
        error: false,
      })
    })
  })
})
