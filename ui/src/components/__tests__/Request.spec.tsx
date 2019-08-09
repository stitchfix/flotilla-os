import * as React from "react"
import { mount, shallow } from "enzyme"
import flushPromises from "flush-promises"
import Request, { Props, ChildProps, RequestStatus } from "../Request"

describe("Request", () => {
  it("calls props.request with props.initialArgs when the component mounts", () => {
    const realRequest = Request.prototype.request
    Request.prototype.request = jest.fn()
    expect(Request.prototype.request).toHaveBeenCalledTimes(0)
    const wrapper = mount(
      <Request
        requestFn={() =>
          new Promise(resolve => {
            resolve()
          })
        }
        initialRequestArgs={{ foo: "bar" }}
      >
        {() => null}
      </Request>
    )
    expect(Request.prototype.request).toHaveBeenCalledTimes(1)
    expect(Request.prototype.request).toHaveBeenCalledWith(
      wrapper.prop("initialRequestArgs")
    )
    Request.prototype.request = realRequest
  })

  it("doesn't call props.request when the component mounts if props.shouldRequestOnMount is false", () => {
    const realRequest = Request.prototype.request
    Request.prototype.request = jest.fn()
    expect(Request.prototype.request).toHaveBeenCalledTimes(0)
    const wrapper = mount(
      <Request
        requestFn={() =>
          new Promise(resolve => {
            resolve()
          })
        }
        initialRequestArgs={{ foo: "bar" }}
        shouldRequestOnMount={false}
      >
        {() => null}
      </Request>
    )
    expect(Request.prototype.request).toHaveBeenCalledTimes(0)
    Request.prototype.request = realRequest
  })

  it("sets state correctly during the request method", async () => {
    const data = "data"
    const onSuccess = jest.fn()
    const successWrapper = shallow(
      <Request
        requestFn={() =>
          new Promise(resolve => {
            resolve(data)
          })
        }
        initialRequestArgs={{ foo: "bar" }}
        onSuccess={onSuccess}
      >
        {(props: ChildProps<any, any>) => null}
      </Request>
    )
    expect(successWrapper.state("requestStatus")).toEqual(
      RequestStatus.NOT_READY
    )
    expect(successWrapper.state("data")).toEqual(null)
    expect(successWrapper.state("isLoading")).toEqual(true)
    expect(successWrapper.state("error")).toEqual(null)
    expect(onSuccess).toHaveBeenCalledTimes(0)
    await flushPromises()
    expect(successWrapper.state("requestStatus")).toEqual(RequestStatus.READY)
    expect(successWrapper.state("data")).toEqual(data)
    expect(successWrapper.state("isLoading")).toEqual(false)
    expect(successWrapper.state("error")).toEqual(null)
    expect(onSuccess).toHaveBeenCalledTimes(1)
    expect(onSuccess).toHaveBeenCalledWith(data)

    const onFailure = jest.fn()
    const err = "err"
    const errorWrapper = shallow(
      <Request
        requestFn={() =>
          new Promise((_, reject) => {
            reject(err)
          })
        }
        initialRequestArgs={{ foo: "bar" }}
        onFailure={onFailure}
      >
        {(props: ChildProps<any, any>) => null}
      </Request>
    )
    expect(errorWrapper.state("requestStatus")).toEqual(RequestStatus.NOT_READY)
    expect(errorWrapper.state("data")).toEqual(null)
    expect(errorWrapper.state("isLoading")).toEqual(true)
    expect(errorWrapper.state("error")).toEqual(null)
    expect(onFailure).toHaveBeenCalledTimes(0)
    await flushPromises()
    expect(errorWrapper.state("requestStatus")).toEqual(RequestStatus.ERROR)
    expect(errorWrapper.state("data")).toEqual(null)
    expect(errorWrapper.state("isLoading")).toEqual(false)
    expect(errorWrapper.state("error")).toEqual(err)
    expect(onFailure).toHaveBeenCalledTimes(1)
    expect(onFailure).toHaveBeenCalledWith(err)
  })
})
